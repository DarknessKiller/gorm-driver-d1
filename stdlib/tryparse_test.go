package stdlib

import (
 "database/sql/driver"
 "testing"
 "time"

 d1 "github.com/kofj/gorm-driver-d1"
)

func TestTryParseTime(t *testing.T) {
 cases := []struct{
   s string
   want bool
 }{
   {"2024-01-02T15:04:05Z", true},
   {"2024-01-02T15:04:05.999999999Z", true},
   {"2024-01-02 15:04:05", true},
   {"2024-01-02 15:04:05.123456789", true},
   {"2024-01-02T15:04:05", true},
   {"2024-01-02", true},
   {"hello", false},
   {"short", false},
   {"", false},
   {"\\u0041\\u0042", false},
   {"2024-13-40 nonsense", false},
 }
 for _, c := range cases {
   _, ok := tryParseTime(c.s)
   if ok != c.want {
     t.Errorf("tryParseTime(%q)=%v want %v", c.s, ok, c.want)
   }
 }
}

func TestRowsNextHeuristic(t *testing.T) {
 // custom column not in old list, should parse as time when parseTime=true
 res := &d1.D1RespQueryResults{
   Columns: []string{"expires_at", "name"},
   Rows: [][]interface{}{{"2024-01-02 15:04:05", "kofj"}},
 }
 rows := &Rows{connId: "test", results: res, parseTime: true}
 dest := make([]driver.Value, 2)
 if err := rows.Next(dest); err != nil {
   t.Fatalf("Next failed: %v", err)
 }
 if _, ok := dest[0].(time.Time); !ok {
   t.Errorf("expires_at should be time.Time, got %T %v", dest[0], dest[0])
 }
 if dest[1].(string) != "kofj" {
   t.Errorf("name should stay string")
 }

 // parseTime=false should keep string
 res2 := &d1.D1RespQueryResults{
   Columns: []string{"expires_at"},
   Rows: [][]interface{}{{"2024-01-02 15:04:05"}},
 }
 rows2 := &Rows{connId: "test", results: res2, parseTime: false}
 dest2 := make([]driver.Value, 1)
 if err := rows2.Next(dest2); err != nil {
   t.Fatalf("Next failed: %v", err)
 }
 if _, ok := dest2[0].(time.Time); ok {
   t.Errorf("with parseTime false, should stay string, got time")
 }
 if dest2[0].(string) != "2024-01-02 15:04:05" {
   t.Errorf("unexpected value")
 }
}

func TestRowsNextUnicodeStillWorks(t *testing.T) {
 // BLOB round-trip via unicode escapes should still become []byte
 res := &d1.D1RespQueryResults{
   Columns: []string{"payload"},
   Rows: [][]interface{}{{"\\u0041\\u0042"}},
 }
 rows := &Rows{connId: "test", results: res, parseTime: true}
 dest := make([]driver.Value, 1)
 if err := rows.Next(dest); err != nil {
   t.Fatalf("Next failed: %v", err)
 }
 b, ok := dest[0].([]byte)
 if !ok || string(b) != "AB" {
   t.Errorf("unicode should become bytes AB, got %T %v", dest[0], dest[0])
 }
}

func TestRowsNextRFC3339CustomCol(t *testing.T){
 res := &d1.D1RespQueryResults{
   Columns: []string{"my_custom_time"},
   Rows: [][]interface{}{{"2024-01-02T15:04:05.999999999Z"}},
 }
 rows := &Rows{connId: "test", results: res, parseTime: true}
 dest := make([]driver.Value, 1)
 rows.Next(dest)
 if _, ok := dest[0].(time.Time); !ok {
   t.Errorf("custom RFC3339 should be time, got %T", dest[0])
 }
}
