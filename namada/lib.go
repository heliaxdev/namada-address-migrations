package namada

import "regexp"

var AddressRegex = regexp.MustCompile(`(atest1.{78})|(p?patest.{76})|(xsktest1.{277})|(xfvktest1.{277})|(d?pktest1.\w+)|(sigtest1\w+)`)
