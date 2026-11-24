import json, os, pathlib, subprocess, tempfile
results=[]
# The lock-file check below needs the files ref backend, not reftable.
env=dict(os.environ, GIT_CONFIG_NOSYSTEM='1', GIT_CONFIG_GLOBAL=os.devnull, GIT_TERMINAL_PROMPT='0', GIT_DEFAULT_REF_FORMAT='files')
def git(cwd,*args,ok=True):
    p=subprocess.run(['git',*args],cwd=cwd,env=env,text=True,capture_output=True)
    if ok and p.returncode: raise AssertionError((args,p.stdout,p.stderr))
    return p
with tempfile.TemporaryDirectory(prefix='cleanup-branches-smoke-') as tmp:
    root=pathlib.Path(tmp); remote=root/'remote.git'; repo=root/'repo'
    git(root,'init','--bare','--initial-branch=main',str(remote))
    git(root,'clone',str(remote),str(repo))
    git(repo,'config','user.name','Fixture'); git(repo,'config','user.email','fixture@example.invalid')
    git(repo,'commit','--allow-empty','-m','base'); git(repo,'push','-u','origin','main')
    base=git(repo,'rev-parse','HEAD').stdout.strip()
    git(repo,'switch','-c','completed'); git(repo,'commit','--allow-empty','-m','completed')
    tip=git(repo,'rev-parse','HEAD').stdout.strip()
    git(repo,'push','-u','origin','completed'); git(repo,'switch','main'); git(repo,'merge','--ff-only','completed'); git(repo,'push','origin','main')
    (repo/'draft.txt').write_text('preserve me\n')
    before=git(repo,'status','--porcelain').stdout
    git(repo,'merge-base','--is-ancestor',tip,'origin/main')
    git(repo,'update-ref','-d','refs/heads/completed',tip)
    git(repo,'config','--remove-section','branch.completed')
    assert git(repo,'show-ref','--verify','refs/heads/completed',ok=False).returncode
    results.append('verified local tip deleted and branch config removed')
    git(repo,'push',f'--force-with-lease=refs/heads/completed:{tip}','origin',':refs/heads/completed')
    assert not git(repo,'ls-remote','--heads','origin','refs/heads/completed').stdout
    results.append('expected-tip remote deletion succeeded and server absence verified')
    git(repo,'switch','-c','moving',base); git(repo,'commit','--allow-empty','-m','first')
    old=git(repo,'rev-parse','HEAD').stdout.strip(); git(repo,'push','-u','origin','moving')
    git(repo,'commit','--allow-empty','-m','unreviewed'); new=git(repo,'rev-parse','HEAD').stdout.strip()
    git(repo,'push','origin','moving'); git(repo,'switch','main')
    assert git(repo,'update-ref','-d','refs/heads/moving',old,ok=False).returncode
    assert git(repo,'rev-parse','moving').stdout.strip()==new
    results.append('local expected-OID deletion rejected moved tip and preserved new commit')
    assert git(repo,'push',f'--force-with-lease=refs/heads/moving:{old}','origin',':refs/heads/moving',ok=False).returncode
    assert new in git(repo,'ls-remote','--heads','origin','refs/heads/moving').stdout
    results.append('remote lease rejected moved tip and preserved server branch')
    assert git(repo,'merge-base','--is-ancestor','moving','main',ok=False).returncode
    git(repo,'branch','-d','moving')
    results.append('branch -d accepted unintegrated work when feature upstream matched: independent integration proof necessary')
    git(repo,'switch','-c','squashed'); (repo/'feature.txt').write_text('feature\n'); git(repo,'add','feature.txt'); git(repo,'commit','-m','feature')
    squash_head=git(repo,'rev-parse','HEAD').stdout.strip(); git(repo,'switch','main'); git(repo,'merge','--squash','squashed'); git(repo,'commit','-m','squash result')
    assert git(repo,'merge-base','--is-ancestor',squash_head,'main',ok=False).returncode
    git(repo,'update-ref','-d','refs/heads/squashed',squash_head)
    results.append('exact-tip deletion supports squash integration without ancestry-based branch -D')
    git(repo,'branch','occupied'); git(repo,'worktree','add',str(root/'occupied'),'occupied')
    assert 'branch refs/heads/occupied' in git(repo,'worktree','list','--porcelain').stdout
    assert git(repo,'branch','-D','occupied',ok=False).returncode
    results.append('worktree inventory identifies occupied branch; porcelain refuses deletion')
    git(repo,'tag','keep-local'); git(repo,'config','fetch.pruneTags','true')
    git(repo,'fetch','--prune','--no-prune-tags','--no-tags','origin','+refs/heads/*:refs/remotes/origin/*')
    git(repo,'show-ref','--verify','refs/tags/keep-local')
    assert git(repo,'status','--porcelain').stdout==before
    assert (repo/'draft.txt').read_text()=='preserve me\n'
    assert git(repo,'branch','--show-current').stdout.strip()=='main'
    results.append('branch-only pruning preserved local tag, dirty file and current branch')
    # Stale tracking refs: the server head is deleted elsewhere (for example by a
    # provider's delete-on-merge), so this clone's own push never updates them.
    other=root/'other'; git(root,'clone',str(remote),str(other))
    for name in ('stale-a','stale-b'):
        git(repo,'switch','-c',name,'main'); git(repo,'commit','--allow-empty','-m',name)
        git(repo,'push','-u','origin',name)
    git(repo,'switch','main')
    tips={n:git(repo,'rev-parse',n).stdout.strip() for n in ('stale-a','stale-b')}
    git(other,'fetch','origin')
    for name in tips: git(other,'push','origin',f':refs/heads/{name}')
    git(repo,'config','--unset','fetch.pruneTags'); git(repo,'config','fetch.prune','false')
    git(repo,'fetch','origin')
    def has(ref): return not git(repo,'show-ref','--verify','--quiet',ref,ok=False).returncode
    listing=git(repo,'branch','-a').stdout
    assert not git(repo,'ls-remote','--heads','origin','refs/heads/stale-a').stdout
    assert has('refs/remotes/origin/stale-a') and has('refs/heads/stale-a')
    assert 'remotes/origin/stale-a' in listing and ' stale-a\n' in listing
    results.append('plain fetch without prune kept stale tracking ref and local head after verified server absence')
    git(repo,'tag','local-only-tag'); git(repo,'config','fetch.pruneTags','true'); git(repo,'config','remote.origin.pruneTags','true')
    git(repo,'fetch','origin','--prune','--no-prune-tags','+refs/heads/*:refs/remotes/origin/*')
    assert not has('refs/remotes/origin/stale-a') and not has('refs/remotes/origin/stale-b')
    assert has('refs/tags/local-only-tag') and has('refs/tags/keep-local')
    assert has('refs/remotes/origin/main')
    results.append('branch-scoped prune removed stale tracking refs and kept tags under fetch.pruneTags and remote pruneTags')
    # Controls: the explicit branch refspec alone already keeps tags; the hazard
    # --no-prune-tags guards against is a plain prune under pruneTags config.
    git(repo,'fetch','origin','--prune','+refs/heads/*:refs/remotes/origin/*')
    assert has('refs/tags/local-only-tag')
    git(repo,'fetch','origin','--prune')
    assert not has('refs/tags/local-only-tag')
    git(repo,'config','--unset','fetch.pruneTags'); git(repo,'config','--unset','remote.origin.pruneTags')
    results.append('controls: branch refspec alone kept the local-only tag; plain prune under pruneTags deleted it')
    lock=repo/'.git'/'refs'/'heads'/'stale-a.lock'; lock.write_text('')
    assert git(repo,'update-ref','-d','refs/heads/stale-a',tips['stale-a'],ok=False).returncode
    assert has('refs/heads/stale-a'); lock.unlink()
    git(repo,'update-ref','-d','refs/heads/stale-b',tips['stale-b'])
    listing=git(repo,'branch','-a').stdout
    assert 'stale-b' not in listing
    assert ' stale-a\n' in listing and 'remotes/origin/stale-a' not in listing
    results.append('locked local deletion failed independently; removed refs left git branch -a, retained local head stayed listed')
print(json.dumps({'git_version':git(None,'--version').stdout.strip(),'results':results,'passed':len(results)},indent=2))
