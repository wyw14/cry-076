package config

import "testing"

func TestProjectIdentity(t *testing.T) {
	if ProjectName != "履历工坊" || ProjectCode != "GO-FE-026" || GlobalSequence != 76 || SourceClassification != "Feature迭代" || SourceReference != "word2.xlsx / Sheet1 / 第 795 行" {
		t.Fatal("project identity metadata changed")
	}
}
