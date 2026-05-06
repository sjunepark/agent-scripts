# 행 정렬 알고리즘: 중복 연도 검증

이 스킬의 핵심. 자동 lookup 대신 **중복된 연도의 숫자가 일치하는지**로 행 정렬을 검증한다.

## 핵심 아이디어

- 4기 FS 파일 → 4기·3기 숫자 포함
- 3기 FS 파일 → 3기·2기 숫자 포함
- 양쪽에 **3기 숫자가 모두 존재** → 같은 행이라면 일치해야 함

따라서 행을 매칭할 때 **계정과목 이름이 아닌 중복 연도의 값**을 fingerprint로 사용한다.

## 단계별 절차

가장 최신 FS(예: 5기 FS)부터 시작해서 점점 오래된 FS로 거슬러 내려가며 합치는 방식.

### Step 1: Master 템플릿 구성

가장 최신 FS를 그대로 master로 채택. 계정과목 순서·이름·실계정 구분(0/1)을 모두 보존. 이게 출력 xlsx의 계정 행 골격이 된다.

### Step 2: 다음 오래된 FS 합치기 (예: 4기 FS → 5기 master에)

5기 FS와 4기 FS는 둘 다 4기 숫자를 가진다. 이 4기 숫자를 일치시킴으로써 행 정렬이 검증된다.

```text
master_idx = 0   # master(5기 FS)의 현재 행 포인터
older_idx  = 0   # 4기 FS의 현재 행 포인터

while older_idx < len(older_fs):
    overlap_master = master[master_idx].value_for_overlap_year   # 4기 숫자 (5기 FS 기준)
    overlap_older  = older[older_idx].value_for_overlap_year     # 4기 숫자 (4기 FS 기준)

    if overlap_master == overlap_older:
        # 정렬 OK. 더 오래된 연도(3기) 값을 master로 가져옴
        master[master_idx].value_for_older_year = older[older_idx].value_for_older_year
        master_idx += 1
        older_idx  += 1
    elif abs(overlap_master - overlap_older) < tolerance and overlap_master != 0:
        # 거의 일치하지만 작은 차이 (예: 단위 변환, 반올림) → 가져오되 플래그
        master[master_idx].value_for_older_year = older[older_idx].value_for_older_year
        master[master_idx].flags["재계산_차이"] = f"4기 비교: master={overlap_master:,} vs older={overlap_older:,}"
        master_idx += 1
        older_idx  += 1
    else:
        # 불일치 → 신규/삭제 계정 가능성. lookahead 시도
        ... (아래 mismatch 처리 참조)
```

### Mismatch 처리 (불일치 발견 시)

자동 추측으로 매칭 시도하지 않는다. 대신:

**Lookahead 1~2 행만**: master의 다음 1~2개 leaf 행 또는 older의 다음 1~2개 leaf 행에 일치하는 값이 있는지 확인.

- master에서 발견 → master에 있는데 older엔 없는 계정. 해당 master 행은 older year를 비움 (신규 계정으로 가정).
- older에서 발견 → older엔 있는데 master엔 없는 계정. 해당 older 행은 무시하거나 master에 행을 삽입할지 판단 필요. 이 경우 삽입은 **사용자에게 확인 필요**.
- 양쪽 다 못 찾음 → `매칭_의심` 플래그 표시 후 일단 그대로 진행. 사용자가 검토.

소계 행(is_leaf=False)은 sum 검증으로 별도 확인이 가능하므로 mismatch 처리에서 lookahead 대상이 아니다 (소계 행은 그냥 양쪽이 같은 위치에 있다고 가정).

### Step 3: 더 오래된 FS도 동일하게 반복

3기 FS → master(이젠 5기·4기·3기 채워진 것)에 합칠 때:
- 중복 연도는 3기 (5기 FS의 3기 칸은 비어있고, 4기 FS의 3기 칸이 채워져 있음 → 후자 사용)
- 새로 가져올 연도는 2기

이 패턴이 매 반복마다 반복됨. 항상 직전 단계의 master에서 가장 오래된 채워진 연도를 overlap year로 사용.

## Overlap year 부재 (gap) 처리

두 FS 파일 사이에 공통 연도가 없으면 fingerprint 검증이 **원천 불가능**하다.

전형적인 케이스:
- BSPL_2023 (4기·3기) + BS_2506 (6기·5기) — 4기와 5기는 한 칸 차이지만 두 파일 모두 4기를 갖고 있지 않음
- 즉 두 파일 사이의 연도가 끊김

### 절차

1. **gap 감지**:
   - 각 파일의 커버 연도를 정렬해놓고 인접 파일끼리 교집합 계산
   - 교집합 비어있으면 gap

2. **사용자에게 먼저 질문**:
   ```
   "X.pdf (커버: 4기·3기)와 Y.pdf (커버: 6기·5기) 사이에
   FY2024(5기) standalone FS 파일이 별도로 있나요?
   없으면 4기↔5기 행 정렬은 위치 매칭만 가능하고
   숫자 cross-check는 불가합니다."
   ```

3. **답이 "없음"이면**:
   - master의 `(매칭_의심)` 플래그를 자동 부착
   - gap 양쪽 두 파일을 위치 기반으로만 정렬
   - 모든 leaf 행에 다음 플래그 추가:
     ```
     "X-Y 사이 fingerprint 검증 불가 (위치만 매칭)"
     ```
   - 단, 명백히 sum check 가 자체적으로 잘 맞으면 (소계와 leaf 합 일치) 일부 신뢰도 회복됨 — 이 경우 플래그를 일부 행에서 제거 가능

4. **답이 "있음"이면**:
   - 중간 파일을 받아 정상 절차로 합침

5. **gap 처리 사실은 최종 보고서에 반드시 명시**:
   ```
   "📍 fingerprint 검증 가능: 2022↔2023
    📍 fingerprint 검증 불가 gap: 2023↔2024 (FY2024 standalone 부재)"
   ```

### Gap이 있어도 부분 검증할 것

위치 기반 매칭이라도 다음은 자체적으로 잘 맞아야 함:
- 소계 sum check — 각 파일 내부에서 자체 검증
- 자산총계 = 부채+자본 — 각 파일 내부에서 자체 검증
- 비가산 검증식 (매출총이익 = 매출-원가 등) — 각 파일 내부에서 자체 검증

이 자체 검증들이 모두 통과하면 위치 매칭이 그래도 신뢰할 만함.

## Tolerance 기준

`tolerance` 는 절대값/상대값 둘 다 고려:
- 절대 차이 ≤ 1 (반올림 오차 후보)
- 또는 상대 차이 ≤ 0.001 (0.1% 미만)

이 기준 안에 들면 일치로 간주하되 `재계산_차이` 에 두 값을 모두 기록 → 사용자가 수기 검증.

`overlap_master == 0` 인 경우는 fingerprint 가치가 낮음 (0인 계정은 많기 때문에 우연 일치 가능). 0끼리 매칭은 받되, 직전·직후 행의 비-0 값 일치 여부도 함께 본다.

## 구현 시 주의

- **소계 행(C=0)과 leaf 행(C=1)은 분리해서 카운트**. master_idx 와 older_idx 는 leaf 행만 이동시키되, 소계 행은 master 구조를 그대로 유지하고 older의 same-name 소계는 자동 매칭한다고 가정해도 무방. 단 매칭 후 **소계 검증값이 어긋나면 그 구간 leaf 매칭을 의심해야 함**.
- 절대 `df.merge(..., on="account")` 같은 string 기반 join을 쓰지 말 것. 사용자가 명시적으로 금지함.
- 매 mismatch 마다 사용자 확인을 받는 건 너무 끊김 → 플래그로 모아서 마지막에 일괄 보고.
