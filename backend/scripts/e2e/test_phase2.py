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
    ok = "OK" if (expect is None or code == expect) else f"!! want {expect}"
    print(f"  {method} {path} -> {code} {ok}")
    if expect is not None and code != expect:
        print("     body:", txt[:300])
        raise SystemExit(1)
    try:
        return json.loads(txt)
    except Exception:
        return txt

def reg(email, **caps):
    return call("POST", "/auth/register", body={"email": email, "password": "supersecret", "displayName": email.split("@")[0], **caps}, expect=201)

print("== setup users ==")
client = reg("client1@ex.com", asClient=True)
frl = reg("frl1@ex.com", asFreelancer=True)
frl2 = reg("frl2@ex.com", asFreelancer=True)
ct, ft, ft2 = client["accessToken"], frl["accessToken"], frl2["accessToken"]

cats = call("GET", "/catalog/categories")["categories"]
dev = next(c for c in cats if c["slug"] == "development")
skills = call("GET", "/catalog/skills?category=development")["skills"]
go_id = next(s["id"] for s in skills if s["slug"] == "go")
pg_id = next(s["id"] for s in skills if s["slug"] == "postgresql")

print("== create + publish project ==")
proj = call("POST", "/projects", ct, {
    "title": "REST API на Go для маркетплейса",
    "description": "Нужен бэкенд с JWT, Postgres, платежами",
    "categoryId": dev["id"],
    "budgetType": "fixed",
    "fixedAmountCents": 15000000,
    "skillIds": [go_id, pg_id],
}, expect=201)
pid = proj["id"]
print("     status:", proj["status"], "skills:", len(proj["skillIds"]))

call("GET", f"/projects/{pid}", None, expect=404)  # draft hidden from anon
call("GET", f"/projects/{pid}", ct, expect=200)    # visible to owner
call("POST", f"/projects/{pid}/proposals", ft, {"coverLetter":"x","bidAmountCents":1}, expect=409)  # not open yet

pub = call("POST", f"/projects/{pid}/publish", ct, expect=200)
print("     published status:", pub["status"], "publishedAt set:", pub["publishedAt"] is not None)

print("== listing / filters ==")
lst = call("GET", "/projects", None, expect=200)
print("     open projects:", lst["total"])
lst = call("GET", f"/projects?category=development&skills={go_id},{pg_id}&budgetType=fixed", None, expect=200)
print("     filtered:", lst["total"])
lst = call("GET", f"/projects?skills={go_id},{pg_id},{next(s['id'] for s in skills if s['slug']=='react')}", None, expect=200)
print("     filtered (needs react too):", lst["total"], "(expect 0)")

print("== proposals ==")
call("POST", f"/projects/{pid}/proposals", ct, {"coverLetter":"x","bidAmountCents":100}, expect=403)  # client can't
p1 = call("POST", f"/projects/{pid}/proposals", ft, {"coverLetter":"Сделаю за 2 недели","bidAmountCents":14000000,"estimatedDays":14}, expect=201)
call("POST", f"/projects/{pid}/proposals", ft, {"coverLetter":"again","bidAmountCents":1}, expect=409)  # dup
p2 = call("POST", f"/projects/{pid}/proposals", ft2, {"coverLetter":"Готов начать сегодня","bidAmountCents":13500000}, expect=201)

proj = call("GET", f"/projects/{pid}", None, expect=200)
print("     proposals_count:", proj["proposalsCount"], "(expect 2)")

props = call("GET", f"/projects/{pid}/proposals", ct, expect=200)["proposals"]
print("     owner sees proposals:", len(props))
call("GET", f"/projects/{pid}/proposals", ft, expect=403)  # not owner

print("== decisions ==")
call("POST", f"/proposals/{p1['id']}/shortlist", ct, expect=200)
call("POST", f"/proposals/{p2['id']}/decline", ct, expect=200)
call("POST", f"/proposals/{p1['id']}/shortlist", ft, expect=403)  # freelancer can't decide
mine = call("GET", "/me/proposals", ft, expect=200)["proposals"]
print("     frl1 proposals:", [(x["status"], x["bidAmountCents"]) for x in mine])
call("POST", f"/proposals/{p1['id']}/withdraw", ft, expect=200)
call("POST", f"/proposals/{p2['id']}/withdraw", ft2, expect=409)  # declined, can't withdraw

print("== validation ==")
call("POST", "/projects", ct, {"title":"abc","description":"x","budgetType":"fixed","fixedAmountCents":100}, expect=400)  # title too short
call("POST", "/projects", ct, {"title":"valid title","description":"x","budgetType":"hourly","hourlyRateMinCents":5000,"hourlyRateMaxCents":1000}, expect=400)  # min>max

print("\nALL PHASE 2 CHECKS PASSED")
