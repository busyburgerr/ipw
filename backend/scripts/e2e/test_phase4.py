import urllib.request, urllib.error, json, subprocess

BASE = "http://localhost:5000/api/v1"

def call(method, path, tok=None, body=None, expect=None):
    h = {"Content-Type": "application/json"}
    if tok:
        h["Authorization"] = "Bearer " + tok
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

print("== setup ==")
client = reg("c4@ex.com", asClient=True)
frl = reg("f4@ex.com", asFreelancer=True)
admin = reg("a4@ex.com", asClient=True)
psql("UPDATE users SET is_admin=true WHERE email='a4@ex.com'")
adm = call("POST", "/auth/login", body={"email":"a4@ex.com","password":"supersecret"}, expect=200)
ct, ft, at = client["accessToken"], frl["accessToken"], adm["accessToken"]

proj = call("POST", "/projects", ct, {"title":"Лендинг и интеграции","description":"React + API",
            "budgetType":"fixed","fixedAmountCents":10000000}, expect=201)
call("POST", f"/projects/{proj['id']}/publish", ct, expect=200)
p = call("POST", f"/projects/{proj['id']}/proposals", ft, {"coverLetter":"Готов","bidAmountCents":10000000}, expect=201)
contract = call("POST", f"/proposals/{p['id']}/accept", ct, expect=201)
cid = contract["id"]
m = call("POST", f"/contracts/{cid}/milestones", ct, {"title":"Весь проект","amountCents":10000000}, expect=201)
mid = m["id"]

print("== fund milestone (stub payment) ==")
call("POST", f"/milestones/{mid}/approve", ct, expect=409)   # not submitted / not funded
pay = call("POST", f"/milestones/{mid}/fund", ct, expect=201)
print("     payment:", pay["provider"], pay["status"], "url:", pay["paymentUrl"][:60])
assert pay["provider"] == "stub" and pay["status"] == "pending"
call("POST", f"/milestones/{mid}/fund", ct, expect=201)      # idempotent: same pending invoice

# milestone still pending until payment confirmed
mstate = call("GET", f"/contracts/{cid}", ct, expect=200)["milestones"][0]
print("     milestone status pre-pay:", mstate["status"])
assert mstate["status"] == "pending"

# simulate payment via dev endpoint
call("POST", f"/dev/payments/{pay['id']}/pay", None, expect=200)
mstate = call("GET", f"/contracts/{cid}", ct, expect=200)["milestones"][0]
print("     milestone status post-pay:", mstate["status"], "fundedAt:", mstate["fundedAt"] is not None)
assert mstate["status"] == "funded"
print("     escrow balance:", psql(f"SELECT COALESCE(SUM(e.amount_cents),0) FROM ledger_entries e JOIN ledger_accounts a ON a.id=e.account_id WHERE a.kind='escrow' AND a.owner_id='{mid}'"))

print("== submit + approve -> release ==")
call("POST", f"/milestones/{mid}/submit", ft, {"deliverableNote":"Задеплоил, доступы в чате"}, expect=200)
appr = call("POST", f"/milestones/{mid}/approve", ct, expect=200)
print("     milestone:", appr["status"], "releasedAt set:", appr["releasedAt"] is not None)
assert appr["status"] == "released"

print("== balances ==")
w = call("GET", "/me/wallet", ft, expect=200)
print("     freelancer wallet:", w)
assert w["availableCents"] == 9000000, w   # 10_000_000 - 10% commission
plat = psql("SELECT COALESCE(SUM(e.amount_cents),0) FROM ledger_entries e JOIN ledger_accounts a ON a.id=e.account_id WHERE a.kind='platform_revenue'")
print("     platform revenue:", plat, "(expect 1000000)")
assert plat == "1000000"
# ledger global invariant: all entries sum to 0
total = psql("SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries")
print("     sum of ALL ledger entries:", total, "(must be 0)")
assert total == "0"

print("== payout ==")
call("POST", "/me/payouts", ft, {"amountCents": 500, "method":"sbp","destination":"79001234567"}, expect=400)  # below min
call("POST", "/me/payouts", ft, {"amountCents": 99999999, "method":"sbp","destination":"79001234567"}, expect=400)  # over balance
po = call("POST", "/me/payouts", ft, {"amountCents": 5000000, "method":"sbp","destination":"79001234567"}, expect=201)
print("     payout:", po["status"], po["amountCents"], "dest:", po["destination"])
w = call("GET", "/me/wallet", ft, expect=200)
print("     wallet after request:", w, "(available down, pending up)")
assert w["availableCents"] == 4000000 and w["pendingPayoutCents"] == 5000000

call("POST", f"/admin/payouts/{po['id']}/process", ct, {"decision":"paid"}, expect=403)  # not admin
done = call("POST", f"/admin/payouts/{po['id']}/process", at, {"decision":"paid","note":"СБП отправлено"}, expect=200)
print("     processed:", done["status"])
w = call("GET", "/me/wallet", ft, expect=200)
print("     wallet after settle:", w)
assert w["availableCents"] == 4000000 and w["pendingPayoutCents"] == 0
total = psql("SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries")
assert total == "0", f"ledger imbalance: {total}"

print("== refund path ==")
proj2 = call("POST", "/projects", ct, {"title":"Второй проект","description":"x","budgetType":"fixed","fixedAmountCents":3000000}, expect=201)
call("POST", f"/projects/{proj2['id']}/publish", ct, expect=200)
p2 = call("POST", f"/projects/{proj2['id']}/proposals", ft, {"coverLetter":"ok","bidAmountCents":3000000}, expect=201)
c2 = call("POST", f"/proposals/{p2['id']}/accept", ct, expect=201)
m2 = call("POST", f"/contracts/{c2['id']}/milestones", ct, {"title":"Этап","amountCents":3000000}, expect=201)
pay2 = call("POST", f"/milestones/{m2['id']}/fund", ct, expect=201)
call("POST", f"/dev/payments/{pay2['id']}/pay", None, expect=200)
call("POST", f"/milestones/{m2['id']}/refund", ct, expect=200)
mstate = call("GET", f"/contracts/{c2['id']}", ct, expect=200)["milestones"][0]
print("     refunded milestone status:", mstate["status"])
assert mstate["status"] == "cancelled"
total = psql("SELECT COALESCE(SUM(amount_cents),0) FROM ledger_entries")
esc = psql(f"SELECT COALESCE(SUM(e.amount_cents),0) FROM ledger_entries e JOIN ledger_accounts a ON a.id=e.account_id WHERE a.kind='escrow' AND a.owner_id='{m2['id']}'")
print("     escrow after refund:", esc, "| global sum:", total)
assert esc == "0" and total == "0"

print("\nALL PHASE 4 CHECKS PASSED")
