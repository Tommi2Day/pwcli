package cmd

const defaultProfileSetName = "default"

// nolint gosec
const defaultProfileSets = `
default:
  profile:
    length: 16
    upper: 1
    lower: 1
    digits: 1
    specials: 1
    first_is_char: true
  special_chars: "!§$%&/()=?-_+<>|#@;:,.[]{}*"
easy:
  profile:
    length: 8
    upper: 1
    lower: 1
    digits: 1
    specials: 0
strong:
  profile:
    length: 48
    upper: 2
    lower: 2
    digits: 2
    specials: 2
    first_is_char: false
  special_chars: "!§$%&/()=?-_+<>|#@;:,.[]{}*"
`
