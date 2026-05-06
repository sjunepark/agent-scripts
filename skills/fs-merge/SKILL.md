---
name: fs-merge
description: "Merge multi-year Korean financial statements (재무제표: BS, PL, 제조원가명세서) from PDFs or spreadsheets into one comparison .xlsx. Use when the user provides two or more Korean FS files and asks to consolidate, compare, 합본, 합치기, combine year-over-year statements, or build a merged workbook."
license: "Personal use"
version: 2
---

# 재무제표 합본 (FS Merge)

## 무엇을 하는 스킬인가

회계·세무 실무자가 제공한 N개의 연도별 재무제표 파일(PDF 또는 xlsx)을 받아, 모든 연도를 한 시트에 정렬해 비교 가능한 합본 xlsx를 만든다. 출력은 `재무제표/손익계산서/제조원가명세서` 3개 섹션이 단일 시트에 들어간 형식.

핵심 제약 두 가지:

1. **계정과목 자동 매칭 금지**. 띄어쓰기·표현 차이로 오차 가능. 행 위치 기반으로 순차 매칭하되, **중복 연도 숫자 일치**로 검증.
2. **불확실한 항목은 별도 플래그 컬럼**에 표시. 사용자가 수기로 검증함.

## 입력 파일 처리

| 형식 | 읽기 방법 |
|------|----------|
| `.xlsx` (정리된 FS) | 사용 가능한 스프레드시트 리더 또는 `openpyxl read_only=True` |
| `.xlsx` (DART 다운로드 raw) | 동일하되 시트가 여러 개일 수 있음 — 시트별로 점검 |
| `.pdf` (텍스트형 — 정상 추출) | `pdftotext -layout`, `pypdf`, 또는 사용 가능한 PDF 텍스트 추출기 |
| `.pdf` (텍스트형 — CID 추출 실패) | **이미지 변환 후 OCR/visual reading** (아래 ⚠ 참조) |
| `.pdf` (스캔본) | `scripts/pdf_to_images.py` 로 페이지를 PNG로 렌더링 → 이미지를 직접 OCR/visual reading으로 읽기 |

PDF가 텍스트형인지 빠르게 판별:
```bash
pdftotext -f 1 -l 1 input.pdf - | head -20
```

세 가지 결과 분기:
- 의미 있는 한국어 텍스트 출력 → `pdftotext` 그대로 사용
- 빈 출력 + 스캔본 의심 → `pdf_to_images.py` 경로
- 빈 출력 + `Syntax Error: Missing language pack for 'Adobe-Korea1' mapping` 또는 `Unknown font tag 'CJK1'` → **CID 인코딩 문제** (다음 ⚠ 참조)

### ⚠ CID 한글 PDF (네이티브 텍스트인데 추출 실패)

세무사·회계법인이 발행한 한글 PDF는 종종 ToUnicode CMap이 누락된 CID 폰트(Adobe-Korea1)를 사용한다. 이 경우 `pdftotext`는 빈 출력 + 위 경고를 낸다. 처리 순서:

1. **백업 시도**: `python -c "import pdfplumber; print(pdfplumber.open('x.pdf').pages[0].extract_text())"` — 일부 케이스에서 동작
2. **PyMuPDF 시도**: `python -c "import fitz; print(fitz.open('x.pdf').load_page(0).get_text())"` — 추가 케이스 동작
3. **모두 실패하면 스캔본과 동일하게 처리**: `pdf_to_images.py` 로 PNG 렌더링 → OCR/visual reading으로 읽기

이 패턴에서는 숫자만 추출되고 한글 계정과목이 사라지는 partial 결과가 자주 나온다. 그런 결과는 **신뢰하지 말고** 이미지 경로로 재시도한다.

스캔본·이미지 변환 PDF에서 숫자를 읽을 때, **자릿수 변환 위험이 있는 셀**은 명시적으로 `OCR_의심` 플래그에 기록한다.

## 실행 환경 / 의존성

- `scripts/build_workbook.py` 는 Python 3 + `openpyxl` 이 필요하다.
- `scripts/pdf_to_images.py` 는 PDF 렌더링이 필요할 때 Python 3 + `PyMuPDF` (`fitz`) 를 사용한다.
- 의존성이 없으면 사용자에게 알리고, 현재 프로젝트/샌드박스에서 패키지 설치가 허용되는 경우에만 설치한다.

## 작업 절차 (6단계)

### Phase 1 — 파일 인벤토리 + 기간 분류

각 입력 파일이 어떤 기수(year)를 커버하는지 정리한다. 이때 **결산기간(period span)도 같이 확인**:

- BS: 잔액일자만 보면 됨 (시점)
- PL/MS: 시작일~종료일 모두 확인. 12개월 미만이면 **중간결산(반기/분기)** 으로 표기

같은 연도가 여러 파일에 나오는 경우, **더 최근에 공시된 FS의 숫자를 채택**한다 (재분류·수정 후 가장 최신 정보).

#### 인벤토리 산출물

다음 형식으로 정리:

```text
파일                  | BS 일자       | PL 기간              | 비고
BSPL_2022.pdf         | 2022-12-31    | 2022-01-01~12-31    | 12M
BSPL_2023.pdf         | 2023-12-31    | 2023-01-01~12-31    | 12M (3기 비교)
                      | 2022-12-31    | 2022-01-01~12-31    |
BS_202506.pdf         | 2025-06-30    | —                   | 6기 중간결산 BS
                      | 2024-12-31    | —                   | 5기 비교
PL_202506.pdf         | —             | 2025-01-01~06-30    | 6M 중간결산 PL
                      | —             | 2024-01-01~12-31    | 12M 5기 비교
```

### Phase 2 — 파일별 추출

각 파일에서 PL · MS · BS 별로 (계정과목, 실계정 0/1, 값 배열) 을 뽑아낸다. 이 단계에서는 합치지 않는다 — 파일별로 raw 데이터만 확보.

스캔본/CID PDF는 이미지 변환 후 visual reading. 한 페이지씩 처리하면서 흐릿한 숫자에 플래그를 단다.

### Phase 3 — Master 템플릿 구성

가장 최근 FS의 계정과목 순서·이름·실계정 구분(0/1)·소계 구조를 그대로 master로 채택. 출력 xlsx의 행 골격이 됨.

### Phase 4 — 순차 매칭 + 중복 연도 검증

**이 스킬의 핵심**. 자세한 알고리즘은 `references/matching_algorithm.md` 참조.

요약:
- master(예: 5기 FS) 와 older(예: 4기 FS) 는 둘 다 4기 숫자를 갖고 있음
- master_idx, older_idx 두 포인터를 leaf 행끼리 동시에 진행
- 매 행마다 4기 값 비교:
  - 일치 → 정렬 OK, older의 3기 값을 master에 채움
  - 작은 차이(±1 또는 0.1% 이하) → `재계산_차이` 플래그 기록 후 진행
  - 큰 불일치 → 1~2행 lookahead 시도, 못 찾으면 `매칭_의심` 플래그 후 진행
- 자동 추측 매칭 금지. 모든 불확실은 플래그로 사용자 검증으로 미룸

#### ⚠ Overlap year 부재 (gap)

두 FS 파일 사이에 공통 연도가 없는 경우(예: 2023 FS 파일 + 2025-06 중간결산 파일, 그 사이 FY2024 standalone 파일 없음) **숫자 fingerprint 검증이 원천 불가능하다**. 처리:

1. **gap 감지되면 사용자에게 먼저 묻기**: "FY2024 단독 FS 파일이 따로 있나요? 없으면 2024 비교칸은 위치 매칭만 가능하고 cross-check 불가합니다."
2. 없다고 하면 **gap 이전 모든 leaf 행에 자동으로 `매칭_의심` 플래그**를 단다 (예: "2023 ↔ 2025/06 사이 FY2024 fingerprint 검증 불가 — 위치만 매칭")
3. gap 처리 사실을 최종 보고서에도 명시

#### 계정과목 변경 (rename / split / merge)

연도별 chart of accounts 변경(예: `매입채무` → `외상매입금`, `현금및현금성자산` → `현금` + `보통예금` 분리)은 자동 매칭하지 말 것. 결정 트리는 `references/account_changes.md` 참조.

### Phase 5 — 검증 합산

- 각 소계 행 = 해당 자식들 합 인지 자체 검증 (`checks` 의 sum 형식)
- **PL 비가산 검증** (반드시 포함):
  - `상품매출원가 = 기초재고 + 당기매입 - 기말재고`
  - `Ⅲ.매출총이익 = Ⅰ.매출액 - Ⅱ.매출원가`
  - `Ⅴ.영업이익 = Ⅲ.매출총이익 - Ⅳ.판매비와관리비`
  - `Ⅷ.소득세차감전이익 = Ⅴ.영업이익 + Ⅵ.영업외수익 - Ⅶ.영업외비용`
  - `Ⅹ.당기순이익 = Ⅷ.소득세차감전이익 - Ⅸ.소득세등`
- BS:
  - `자산총계 = 부채총계 + 자본총계` (구조상 자동으로 맞음)
  - `유동자산 = 당좌자산 + 재고자산`, `비유동자산 = 투자+유형+무형+기타`
- 제조업의 경우: PL `당기제품제조원가` (Ⅱ.매출원가 하위) = MS `Ⅺ.당기제품제조원가` (Ref 섹션에서 차이 표시)

비가산 검증식은 `checks[].terms` 필드로 부호 있는 선형 결합 표현 (`references/output_layout.md` 참조).

#### 자본 롤포워드 검증 — 조건부

`기말 이익잉여금 = 시초 + 당기순이익 ± 기타자본변동` 검증은 자본 섹션이 다음 구조일 때만 수행:
- 자본금 / 이익잉여금이 별도 행으로 분리되어 있음
- 시초·기말 잔액이 명확

개인사업자·소규모법인이 자본금 단일 행으로 처리하는 경우(자본+이익+인출금 lump-sum) **이 검증은 skip**한다. 인출금 정보가 BS엔 없고 시산표에만 있어서 재구성 불가능.

검증 실패는 사용자에게 보고하되 출력은 일단 생성한다 (사용자가 결과를 보고 판단).

### Phase 6 — 출력 빌드

`scripts/build_workbook.py` 에 JSON 입력으로 데이터를 넘겨 xlsx를 생성한다.

```bash
python scripts/build_workbook.py /tmp/payload.json <output-dir>/FS_합본.xlsx
```

JSON schema는 빌드 스크립트 docstring 참조. 핵심 필드:

- `fiscal_year_ends`: 결산기말일 ISO 문자열 배열 (오래된 것부터)
- `period_notes` (선택): `fiscal_year_ends`와 평행한 배열. 12개월이 아닌 기간엔 `"6M interim"` 등 짧은 라벨. 12M 정상이면 `null`. 헤더 아래 별도 행으로 표시됨.
- `statements.PL` / `MS` / `BS`: 행 객체 배열. 각 행은 `account`, `is_leaf`, `values`, `flags`
- `statements.BS` 의 행은 `group` 필드 추가 가능 (사용자 지시상 비워두는 게 default)
- 행 객체에 `is_section_header: true` 가 있으면 실계정(C) 컬럼을 비워둠 (예: `[자산]`/`[부채]`/`[자본]` 시각적 구분 행)
- `checks`: 소계 검증. 두 가지 형식 지원:
  - `child_indices: [a, b, c]` — 모두 양수 합 (legacy)
  - `terms: [{idx, sign}, ...]` — 부호 있는 선형 결합 (비가산 검증용)
- `refs`: 명세서 간 cross-reference 검증

소계는 명시적으로 `is_leaf: false` 로 표시. **`is_leaf: true` 인 행만 master/older 매칭의 포인터 진행 대상**이다.

비제조업(MS 없음)의 경우 `"MS": []` 로 빈 배열 전달 — build script가 섹션 자체를 스킵.

레이아웃 상세는 `references/output_layout.md` 참조.

## 출력 후 보고

xlsx 파일을 사용자가 지정했거나 현재 실행 환경에서 접근 가능한 출력 경로에 저장하고, 해당 환경의 파일 전달/첨부 방식으로 사용자에게 전달한다.
함께 다음을 요약 보고:

- 처리한 파일 N개와 각각 커버하는 기수·기간 (12M / 6M / 3M 표기)
- 출력 연도 범위
- **Overlap-year fingerprint 검증 가능 구간 / 검증 불가 gap 구간**을 명시 (예: "2022↔2023 검증 OK, 2023↔2024 gap (FY2024 standalone 파일 없음)")
- 매칭 단계에서 발견한 이슈 (각 플래그 종류별 건수와 대표 항목)
- 합산 검증 실패 항목 (있으면)
- 중간결산(non-12M)이 있으면 PL 비교 시 주의사항 별도 강조

플래그가 많아도 그대로 보고. 사용자가 수기 검증할 거라고 명시적으로 말함.

## 절대 하지 말 것

- 계정과목 이름으로 자동 lookup/merge (string 기반 join)
- 불일치를 휴리스틱으로 임의 해결 (예: "비슷한 이름이니까 같은 거겠지")
- 사용자에게 묻지 않고 행 삽입/삭제
- BS Group 컬럼(NOA/NWC/Capex/Debt/Tax 등) 자동 분류 시도 — 사용자가 수기로 함
- 출력 xlsx에 색상·테두리·기타 스타일 추가 (사용자가 minimal 선호)
- 단위 통일 시도 — 원본이 천원/백만원이면 그대로 유지하고 사용자에게 알림
- **중간결산 PL을 12M PL과 단순 비교** — 컬럼 헤더에 기간 명시 필수
- **부분 추출된 한글 PDF 결과를 그대로 사용** — 한글이 사라지고 숫자만 남으면 이미지 경로로 재시도

## DART 링크는 무시

DART URL을 받아도 직접 fetch 하지 않는다. 사용자에게 PDF/xlsx 다운로드 후 첨부해달라고 요청.

## v2 변경사항 (v1 대비)

- CID 한글 PDF 추출 실패 케이스의 fallback 절차 명시
- Overlap-year gap 발생 시 처리 절차 추가 (`references/matching_algorithm.md`)
- 중간결산(non-12M) 감지 및 `period_notes` schema 추가
- 비가산 PL 검증식(매출총이익/영업이익 등) 표준 체크리스트 명시
- 자본 롤포워드 검증의 조건부 적용 명시
- 계정과목 변경(rename/split/merge) 결정 트리 (`references/account_changes.md` 신규)
- `checks[].terms` 부호 있는 선형 결합 schema 추가
- `is_section_header` 행 옵션 추가 (`[자산]`/`[부채]`/`[자본]` 등)
- 차감계정(감가상각누계액 등) 음수 입력 규약 명시 (`references/output_layout.md`)
