import urllib.request, urllib.error, json, subprocess

BASE = "http://localhost:5000/api/v1"

def call(method, path, tok=None, body=None, expect=None):
    h = {"Content-Type": "application/json"}
    if tok: h["Authorization"] = "Bearer " + tok
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE + path, method=method, data=data, headers=h)
    try:
        r = urllib.request.urlopen(req); code, txt = r.code, r.read().decode()
    except urllib.error.HTTPError as e:
        code, txt = e.code, e.read().decode()
    tag = "OK" if (expect is None or code == expect) else f"!! want {expect}"
    print(f"  {method} {path} -> {code} {tag}")
    if expect is not None and code != expect:
        print("     ", txt[:300]); raise SystemExit(1)
    try: return json.loads(txt)
    except Exception: return txt

def reg(email, **caps):
    return call("POST", "/auth/register", body={"email": email, "password": "supersecret",
               "displayName": email.split("@")[0], **caps}, expect=201)

def psql(sql):
    return subprocess.run(["docker","exec","ipw-postgres-1","psql","-U","ipw","-d","ipw","-tA","-c",sql],
                          capture_output=True, text=True).stdout.strip()

def run_contract(client_tok, frl_tok, amount, title):
    proj = call("POST", "/projects", client_tok, {"title": title, "description": "x",
                "budgetType": "fixed", "fixedAmountCents": amount}, expect=201)
    call("POST", f"/projects/{proj['id']}/publish", client_tok, expect=200)
    p = call("POST", f"/projects/{proj['id']}/proposals", frl_tok, {"coverLetter": "ok", "bidAmountCents": amount}, expect=201)
    ct = call("POST", f"/proposals/{p['id']}/accept", client_tok, expect=201)
    m = call("POST", f"/contracts/{ct['id']}/milestones", client_tok, {"title": "All", "amountCents": amount}, expect=201)
    pay = call("POST", f"/milestones/{m['id']}/fund", client_tok, expect=201)
    call("POST", f"/dev/payments/{pay['id']}/pay", None, expect=200)
    return ct["id"], m["id"]

print("== setup ==")
c = reg("c5@ex.com", asClient=True); C = c["accessToken"]
f = reg("f5@ex.com", asFreelancer=True); F = f["accessToken"]
call("PUT", "/me/freelancer-profile", F, {"headline": "Dev", "availability": "available"}, expect=200)
call("PUT", "/me/client-profile", C, {"companyName": "Acme"}, expect=200)
adm = reg("a5@ex.com", asClient=True)
psql("UPDATE users SET is_admin=true WHERE email='a5@ex.com'")
A = call("POST", "/auth/login", body={"email": "a5@ex.com", "password": "supersecret"}, expect=200)["accessToken"]

print("== reviews (double-blind) ==")
cid, mid = run_contract(C, F, 4000000, "Проект для отзыва")
call("POST", f"/contracts/{cid}/reviews", C, {"rating": 5, "comment": "x"}, expect=409)  # not completed yet
call("POST", f"/milestones/{mid}/submit", F, {"deliverableNote": "done"}, expect=200)
call("POST", f"/milestones/{mid}/approve", C, expect=200)
call("POST", f"/contracts/{cid}/complete", C, expect=200)

r1 = call("POST", f"/contracts/{cid}/reviews", C, {"rating": 5, "comment": "Отличный подрядчик"}, expect=201)
print("     client review published?", r1["published"], "(expect False - blind)")
assert r1["published"] is False
# freelancer sees only own review so far
vis = call("GET", f"/contracts/{cid}/reviews", F, expect=200)["reviews"]
print("     freelancer sees", len(vis), "review(s) before submitting own")
assert len(vis) == 0
call("POST", f"/contracts/{cid}/reviews", C, {"rating": 3}, expect=409)  # already reviewed

r2 = call("POST", f"/contracts/{cid}/reviews", F, {"rating": 4, "comment": "Хороший заказчик"}, expect=201)
print("     after both submit, freelancer review published?", r2["published"])
assert r2["published"] is True
vis = call("GET", f"/contracts/{cid}/reviews", F, expect=200)["reviews"]
print("     both now visible:", len(vis))
assert len(vis) == 2

fr = psql("SELECT rating_avg, rating_count FROM freelancer_profiles WHERE user_id=(SELECT id FROM users WHERE email='f5@ex.com')")
print("     freelancer profile rating:", fr, "(expect 5.00 | 1)")
pub = call("GET", f"/users/{f['user']['id']}/reviews", expect=200)["reviews"]
assert len(pub) == 1 and pub[0]["rating"] == 5

print("== disputes ==")
cid2, mid2 = run_contract(C, F, 6000000, "Спорный проект")
call("POST", f"/milestones/{mid2}/submit", F, {"deliverableNote": "мой вариант"}, expect=200)
call("POST", f"/contracts/{cid2}/disputes", A, {"milestoneId": mid2, "reason": "не участник"}, expect=403)
d = call("POST", f"/contracts/{cid2}/disputes", F, {"milestoneId": mid2, "reason": "Заказчик не отвечает, работа сделана"}, expect=201)
print("     dispute:", d["status"])
call("POST", f"/contracts/{cid2}/disputes", C, {"milestoneId": mid2, "reason": "duplicate"}, expect=409)  # one live per contract
cstat = psql(f"SELECT status FROM contracts WHERE id='{cid2}'")
print("     contract status:", cstat, "(expect disputed)")
assert cstat == "disputed"
call("POST", f"/milestones/{mid2}/approve", C, expect=409)  # can't approve during dispute (milestone submitted, contract not active)

open_d = call("GET", "/admin/disputes", A, expect=200)["disputes"]
print("     arbiter queue:", len(open_d))
call("GET", "/admin/disputes", C, expect=403)
call("POST", f"/admin/disputes/{d['id']}/claim", A, expect=200)
res = call("POST", f"/admin/disputes/{d['id']}/resolve", A, {"outcome": "freelancer", "note": "Работа принята арбитром"}, expect=200)
print("     resolved:", res["status"])
assert res["status"] == "resolved_freelancer"
mstat = psql(f"SELECT status FROM milestones WHERE id='{mid2}'")
print("     milestone status:", mstat, "(expect released)")
assert mstat == "released"

# freelancer got paid (minus commission) for BOTH contracts: 4M*0.9 + 6M*0.9 = 9,000,000
bal = call("GET", "/me/wallet", F, expect=200)
print("     freelancer balance:", bal["availableCents"], "(expect 9000000)")
assert bal["availableCents"] == 9000000
assert psql("SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries") == "0"

print("\nALL PHASE 5 CHECKS PASSED")
