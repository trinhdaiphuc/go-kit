package main

//go:generate msgp

// SmallStruct: 5 fields
type SmallStruct struct {
	Field1 string `redis:"field1" json:"field1"`
	Field2 int    `redis:"field2" json:"field2"`
	Field3 bool   `redis:"field3" json:"field3"`
	Field4 string `redis:"field4" json:"field4"`
	Field5 int    `redis:"field5" json:"field5"`
}

// MediumStruct: 10 fields
type MediumStruct struct {
	SmallStruct `msg:",inline"`
	Field6      string  `redis:"field6" json:"field6"`
	Field7      int     `redis:"field7" json:"field7"`
	Field8      bool    `redis:"field8" json:"field8"`
	Field9      float64 `redis:"field9" json:"field9"`
	Field10     string  `redis:"field10" json:"field10"`
}

// LargeStruct: 25 fields (Flat)
type LargeStruct struct {
	MediumStruct `msg:",inline"`
	Field11      string `redis:"field11" json:"field11"`
	Field12      int    `redis:"field12" json:"field12"`
	Field13      bool   `redis:"field13" json:"field13"`
	Field14      string `redis:"field14" json:"field14"`
	Field15      int    `redis:"field15" json:"field15"`
	Field16      string `redis:"field16" json:"field16"`
	Field17      int    `redis:"field17" json:"field17"`
	Field18      bool   `redis:"field18" json:"field18"`
	Field19      string `redis:"field19" json:"field19"`
	Field20      int    `redis:"field20" json:"field20"`
	Field21      string `redis:"field21" json:"field21"`
	Field22      int    `redis:"field22" json:"field22"`
	Field23      bool   `redis:"field23" json:"field23"`
	Field24      string `redis:"field24" json:"field24"`
	Field25      int    `redis:"field25" json:"field25"`
}

// Nested parts
type SubStruct struct {
	SubField1 string `json:"sub_field1"`
	SubField2 int    `json:"sub_field2"`
}

// LargeNestedStruct: 25 fields total equivalent (approx), with nesting
// 5 flat fields + 10 nested structs (each with 2 fields) = 25 data points
type LargeNestedStruct struct {
	Field1   string    `json:"field1"`
	Field2   int       `json:"field2"`
	Field3   bool      `json:"field3"`
	Field4   string    `json:"field4"`
	Field5   int       `json:"field5"`
	Nested1  SubStruct `json:"nested1"`
	Nested2  SubStruct `json:"nested2"`
	Nested3  SubStruct `json:"nested3"`
	Nested4  SubStruct `json:"nested4"`
	Nested5  SubStruct `json:"nested5"`
	Nested6  SubStruct `json:"nested6"`
	Nested7  SubStruct `json:"nested7"`
	Nested8  SubStruct `json:"nested8"`
	Nested9  SubStruct `json:"nested9"`
	Nested10 SubStruct `json:"nested10"`
}
