package main

import "encoding/json"

// jsonMarshal exists to allow forcing marshal failures in unit tests.
// Production code uses encoding/json.Marshal.
var jsonMarshal = json.Marshal
