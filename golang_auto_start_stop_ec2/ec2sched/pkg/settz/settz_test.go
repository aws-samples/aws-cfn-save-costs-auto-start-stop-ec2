package settz

import (
	"testing"
)

func TestSetRegion(t *testing.T) {

	t.Run("Usecase 1", func(t *testing.T) {
		result, _ := SetRegion("")
		expected_res := "UTC"

		if result != expected_res {
			t.Errorf("SetRegion(\"\") = %s; expected %s", result, expected_res)
		}
	})

	t.Run("Usecase 2", func(t *testing.T) {
		result, _ := SetRegion("America/Los_Angeles")
		expected_res := "America/Los_Angeles"

		if result != expected_res {
			t.Errorf("SetRegion(\"America/Los_Angeles\") = %s; expected %s", result, expected_res)
		}
	})

	t.Run("Usecase 3", func(t *testing.T) {
		_, err := SetRegion("America/Los_Angelos")
		expected_err := "unknown time zone America/Los_Angelos"

		if err.Error() != expected_err {
			t.Errorf("SetRegion(\"America/Los_Angelos\") = %s; expected %s", err.Error(), expected_err)
		}
	})
}
