# End-to-end flow scripts

Plain `urllib` scripts that drive the API through a full marketplace scenario.
They double as living API documentation.

Each script assumes a **fresh database** and a running API on `:5000` with the
stub payment provider (no `LAVA_API_KEY`).

```sh
# from backend/
docker compose up -d
docker exec ipw-postgres-1 psql -U ipw -d ipw -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
go run ./cmd/api &            # re-applies migrations + seeds

python scripts/e2e/test_phase2.py   # projects + proposals
python scripts/e2e/test_phase3.py   # contracts + milestones
python scripts/e2e/test_phase4.py   # ledger / escrow / payments / payouts
python scripts/e2e/test_phase5.py   # reviews + disputes
```

`test_phase4` and `test_phase5` shell out to `docker exec ipw-postgres-1 psql`
to assert ledger balances directly.
