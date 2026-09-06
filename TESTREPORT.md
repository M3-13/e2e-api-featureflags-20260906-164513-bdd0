VERDICT: BUGS_FOUND

- **Title:** Test-Build bricht wegen doppelt deklarierter Hilfsfunktion `doRequest`
- **Symptom:** `go test ./...` schlägt fehl; das Paket `featureflags/internal/httpapi` wird nicht gebaut, daher laufen die HTTP-API-Handler-Tests gar nicht. Damit ist AC-10 (`go test ./...` läuft grün) verletzt.
- **Repro:** `go test ./...` im Projektstamm ausführen.
- **Evidence:**
  ```
  FAIL	featureflags/internal/httpapi [build failed]
  internal\httpapi\update_test.go:25:6: doRequest redeclared in this block
  	internal\httpapi\flags_test.go:49:6: other declaration of doRequest
  ```
- **Suspected file(s):** `internal/httpapi/update_test.go` und `internal/httpapi/flags_test.go` — beide deklarieren im selben Paket `httpapi` die Funktion `doRequest`; eine der beiden Deklarationen muss umbenannt oder in eine gemeinsame Test-Helferdatei ausgelagert werden.
- **Severity:** high