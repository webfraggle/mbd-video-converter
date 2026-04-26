package ui

import "strconv"

func atoi(s string) (int, error)         { return strconv.Atoi(s) }
func atof(s string) (float64, error)     { return strconv.ParseFloat(s, 64) }
