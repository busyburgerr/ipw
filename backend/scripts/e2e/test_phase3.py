import urllib.request, urllib.error, json

BASE = "http://localhost:5000/api/v1"

def call(method, path, tok=None, body=None, expect=None):
    h = {"Content-Type": "application/json"}
    if tok:
        h["Authorization"] = "Bearer " + tok
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE + path, method=method, data=data, headers=h)
    try:
        resp = urllib.request.urlopen(req)
        code, txt = resp.code, resp.read().decode()
    except urllib.error.HTTPError as e:
        code, txt = e.code, e.read().decode()
    tag = "OK" if (expect is None or code == expect) else f"!! want {expect}"
    print(f"  {method} {path} -> {code} {tag}")
    if expect is not None and code != expect:
        print("     body:", txt[:300]); raise SystemExit(1)
    try:
        return json.loads(txt)
    except Exception:
        return txt

def reg(email, **caps):
    return call("POST", "/auth/register", body={"email": email, "password": "supersecret",
               "displayName": email.split("@")[0], **caps}, expect=201)

print("== setup ==")
client = reg("c3@ex.com", asClient=True)
f1 = reg("f3a@ex.com", asFreelancer=True)
f2 = reg("f3b@ex.com", asFreelancer=True)
ct, t1, t2 = client["accessToken"], f1["accessToken"], f2["accessToken"]

proj = call("POST", "/projects", ct, {
    "title": "Мобильное приложение под ключ",
    "description": "Flutter + Go backend",
    "budgetType": "fixed", "fixedAmountCents": 50000000,
}, expect=201)
pid = proj["id"]
call("POST", f"/projects/{pid}/publish", ct, expect=200)
p1 = call("POST", f"/projects/{pid}/proposals", t1, {"coverLetter": "Портфолио внутри", "bidAmountCents": 48000000, "estimatedDays": 40}, expect=201)
p2 = call("POST", f"/projects/{pid}/proposals", t2, {"coverLetter": "Дешевле", "bidAmountCents": 40000000}, expect=201)

print("== accept proposal -> contract ==")
call("POST", f"/proposals/{p2['id']}/accept", t1, expect=403)          # freelancer can't accept
contract = call("POST", f"/proposals/{p1['id']}/accept", ct, expect=201)
cid = contract["id"]
print("     contract:", contract["type"], contract["agreedAmountCents"], contract["status"])
assert contract["agreedAmountCents"] == 48000000 and contract["type"] == "fixed"

call("POST", f"/proposals/{p1['id']}/accept", ct, expect=409)          # already accepted
proj = call("GET", f"/projects/{pid}", ct, expect=200)
print("     project status:", proj["status"], "(expect in_progress)")
mine2 = call("GET", "/me/proposals", t2, expect=200)["proposals"]
print("     other proposal status:", mine2[0]["status"], "(expect declined)")

print("== visibility ==")
call("GET", f"/contracts/{cid}", t1, expect=200)   # freelancer party
call("GET", f"/contracts/{cid}", t2, expect=403)   # outsider
mc = call("GET", "/me/contracts", ct, expect=200)["contracts"]
print("     client contracts:", len(mc))

print("== milestones ==")
call("POST", f"/contracts/{cid}/milestones", ct, {"title": "Дизайн + прототип", "amountCents": 20000000}, expect=201)
m2 = call("POST", f"/contracts/{cid}/milestones", ct, {"title": "Разработка", "amountCents": 28000000}, expect=201)
call("POST", f"/contracts/{cid}/milestones", ct, {"title": "Слишком много", "amountCents": 1}, expect=400)  # exceeds agreed
call("POST", f"/contracts/{cid}/milestones", t1, {"title": "x", "amountCents": 1}, expect=403)  # freelancer can't add

ms = call("GET", f"/contracts/{cid}", ct, expect=200)["milestones"]
m1id = ms[0]["id"]; m2id = m2["id"]
print("     milestones:", [(m["sequence"], m["status"], m["amountCents"]) for m in ms])

print("== milestone lifecycle (m1) ==")
call("POST", f"/milestones/{m1id}/submit", t1, expect=409)             # not funded
call("POST", f"/milestones/{m1id}/fund", ct, expect=200)
call("POST", f"/milestones/{m1id}/submit", ct, {"deliverableNote": "x"}, expect=403)  # client can't submit
call("POST", f"/milestones/{m1id}/submit", t1, {"deliverableNote": "Готово, ссылка на Figma"}, expect=200)
call("POST", f"/milestones/{m1id}/request-changes", ct, expect=200)    # back to funded
call("POST", f"/milestones/{m1id}/submit", t1, {"deliverableNote": "Поправил"}, expect=200)
m1 = call("POST", f"/milestones/{m1id}/approve", ct, expect=200)
print("     m1:", m1["status"], "fundedAt/submittedAt/approvedAt set:",
      all(m1[k] for k in ("fundedAt", "submittedAt", "approvedAt")))

print("== complete contract ==")
call("POST", f"/contracts/{cid}/complete", ct, expect=409)             # m2 still pending
call("POST", f"/milestones/{m2id}/fund", ct, expect=200)
call("POST", f"/milestones/{m2id}/submit", t1, {"deliverableNote": "Код в репозитории"}, expect=200)
call("POST", f"/milestones/{m2id}/approve", ct, expect=200)
done = call("POST", f"/contracts/{cid}/complete", ct, expect=200)
print("     contract status:", done["status"], "endedAt set:", done["endedAt"] is not None)

print("\nALL PHASE 3 CHECKS PASSED")
