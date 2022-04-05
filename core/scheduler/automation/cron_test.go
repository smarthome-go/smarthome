package automation

import (
	"testing"
)

func TestGenerateCronExpression(t *testing.T) {
	tables := []struct {
		Hour        uint8
		Minute      uint8
		Days        []uint8
		Result      string
		HumanResult string
	}{
		{4, 5, []uint8{0, 1, 2, 3, 4, 5, 6}, "5 4 * * *", "At 04:05 AM"},
		{22, 0, []uint8{1, 2, 3, 4, 5}, "0 22 * * 1,2,3,4,5", "At 10:00 PM, only on Monday, Tuesday, Wednesday, Thursday, and Friday"},
		{4, 5, []uint8{0}, "5 4 * * 0", "At 04:05 AM, only on Sunday"},
	}
	for _, table := range tables {
		got, err := GenerateCronExpression(table.Hour, table.Minute, table.Days)
		if err != nil {
			t.Errorf("Test Failed: failed to generate cron expression")
			return
		}
		if got != table.Result {
			t.Errorf("Test (H: %d M: %d DAYS: %v) Failed: OUTPUT: %s", table.Hour, table.Minute, table.Days, table.Result)
			return
		}
		if !IsValidCronExpression(got) {
			t.Errorf("Test (H: %d M: %d DAYS: %v) Failed: output may be an invalid cron expression: %s", table.Hour, table.Minute, table.Days, table.Result)
			return
		}
		days, err := GetDaysFromCronExpression(got)
		if err != nil {
			t.Error(err.Error())
			return
		}
		for index, day := range days {
			if day != table.Days[index] {
				t.Errorf("Test (H: %d M: %d DAYS: %v) Failed: output may contain invalid days: %s", table.Hour, table.Minute, table.Days, table.Result)
			}
		}
		human, err := generateHumanReadableCronExpression(got)
		if err != nil {
			t.Error(err.Error())
			return
		}
		if human != table.HumanResult {
			t.Errorf("Test (H: %d M: %d DAYS: %v) Failed: human readable output should be: %s but got: %s", table.Hour, table.Minute, table.Days, table.HumanResult, human)
			return
		}
	}
}
