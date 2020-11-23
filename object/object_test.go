package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}
	if hello1.Hash() != hello2.Hash() {
		t.Errorf("strings with same content have different hash keys")
	}
	if diff1.Hash() != diff2.Hash() {
		t.Errorf("strings with same content have different hash keys")
	}
	if hello1.Hash() == diff1.Hash() {
		t.Errorf("strings with different content have same hash keys")
	}
}
