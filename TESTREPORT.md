VERDICT: BUGS_FOUND

Der Build (`go build ./...`) ist sauber, aber der Testlauf `go test ./...` schlägt fehl (exit 1). Die `recovered from panic`-Zeilen stammen aus dem Recover-Middleware-Test und sind dort erwartetes Logging; der eigentliche Befund ist ein API-Vertragsverstoß, der neun API-Tests bereits im Setup scheitern lässt.

- **Title**: POST /flags verlangt fälschlich das Pflichtfeld `enabled` und bricht dadurch die Setup-Phase der API-Tests
- **Symptom**: Clients, die ein Flag ohne `enabled` anlegen (nur `key` bzw. weitere optionale Felder), erhalten 400 `{"error":"enabled is required"}` statt 201. Dadurch scheitern `TestListFlags`, `TestGetFlag`, `TestUpdateFlag`, `TestDeleteFlag`, `TestEvaluateFlag`, `TestLoggingMiddleware`, `TestFlagLimit` und `TestStoreDoesNotPersistUser` bereits im Setup, sodass die eigentlichen Handler nie geprüft werden. AC-10 (`go test ./...` läuft grün) ist damit verletzt.
- **Repro**: `go test ./...` im Repository ausführen; die Tests rufen POST /flags mit Body ohne `enabled` auf, z. B. `{"key":"l1_1788708427974818900_4"}`.
- **Evidence**:
  ```
  --- FAIL: TestListFlags (0.00s)
      flags_test.go:162: setup: create "l1_1788708427974818900_4" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestGetFlag (0.00s)
      flags_test.go:183: setup: create "get_1788708427974818900_5" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestUpdateFlag (0.00s)
      flags_test.go:209: setup: create "upd_1788708427974818900_6" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestDeleteFlag (0.00s)
      flags_test.go:244: setup: create "del_1788708427974818900_7" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestEvaluateFlag (0.00s)
      flags_test.go:270: setup: create "ev_1788708427974818900_8" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestLoggingMiddleware (0.00s)
      flags_test.go:377: setup: create "log_1788708428010875000_10" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestFlagLimit (0.00s)
      flags_test.go:410: setup: create "lim_1788708428011398100_11" returned 400 (body {"error":"enabled is required"}
          )
  --- FAIL: TestStoreDoesNotPersistUser (0.00s)
      flags_test.go:425: setup: create "priv_1788708428011398100_12" returned 400 (body {"error":"enabled is required"}
          )
  FAIL
  FAIL	featureflags/internal/httpapi	0.469s
  ```
- **Suspected file(s)**: `internal/httpapi/create.go` — dort wird `req.Enabled == nil` mit 400 beantwortet; gemäß AC-01/AC-02 ist `enabled` nicht als Pflichtfeld spezifiziert, sodass Tests, die ohne `enabled` anlegen, fehlschlagen.
- **Severity**: high