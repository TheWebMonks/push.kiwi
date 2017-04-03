package main

import (
    "fmt"
)

const (
    Proto = "0"
    Major = "0"
    Minor = "2beta"
)

func MajorMinor() string {
    return fmt.Sprintf("%s.%s", Major, Minor)
}

func Full() string {
    return fmt.Sprintf("%s.%s.%s", Proto, Major, Minor)
}

func Compat(client string, server string) bool {
    return client == server
}
