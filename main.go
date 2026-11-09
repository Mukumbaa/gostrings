package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
	"github.com/Mukumbaa/goargv"
)

func is_printable(r rune, include_all_whitespace bool, unicode byte) bool{
	if (r >= 32 && r <= 126) {
		return true
	}
	if (include_all_whitespace) {
		return (r >= 9 && r <= 13)
	}
	if (unicode == 'l' || unicode == 'e' || unicode == 'x' || unicode == 'h') && r > 127 && r != utf8.RuneError {
		return true
	}

	return false
}

func create_rune_from_block(block []byte, encoding byte) rune{
	switch encoding{
	case 's', 'S':
		return rune(block[0])
	case 'l':
		return rune(binary.LittleEndian.Uint16(block))
	case 'b':
		return rune(binary.BigEndian.Uint16(block))
	case 'L':
		return rune(binary.LittleEndian.Uint32(block))
	case 'B':
		return rune(binary.BigEndian.Uint32(block))
	}
	return rune(block[0])
}

func main(){
	var min_len = 4
	var source = []string{}
	var print_file_name = false
	var include_all_whitespace = false
	var radix byte = ' '
	var radix_o = false
	var custom_separator = "\n"
	var encoding byte = 's'
	var unicode byte = 'd'
	var byte_to_take = 1

	p := goargv.New("gostrings")	
	p.BindRest("source", &source, true, "", "Source files")
	p.BindInt("n","bytes",&min_len, false, "", "Minimum bytes in a string")
	p.BindChar("t", "radix", &radix, false, "d,x,o", "Print offset")
	p.SetChoiceHelp("radix", "d", "decimal offset")
	p.SetChoiceHelp("radix", "x", "hexadecimal offset")
	p.SetChoiceHelp("radix", "o", "octal offset")
	p.BindChar("e", "encoding", &encoding, false, "s,S,b,B,l,L", "Select the character encoding of the strings that are to be found")
	p.SetChoiceHelp("encoding", "s", "single-7-bit-byte characters")
	p.SetChoiceHelp("encoding", "S", "single-8-bit-byte characters")
	p.SetChoiceHelp("encoding", "b", "16-bit bigendian")
	p.SetChoiceHelp("encoding", "B", "32-bit bigendian")
	p.SetChoiceHelp("encoding", "l", "16-bit littleendian")
	p.SetChoiceHelp("encoding", "L", "32-bit littleendian")
	p.BindChar("U", "unicode", &unicode, false, "d,i,l,e,x,h", "Controls the display of UTF-8 encoded multibyte characters in strings")
	p.SetChoiceHelp("unicode", "d", "default (rely on --encoding)")
	p.SetChoiceHelp("unicode", "i", "treat multibyte as invalid")
	p.SetChoiceHelp("unicode", "l", "display in current locale")
	p.SetChoiceHelp("unicode", "e", "display as \\uXXXX escape")
	p.SetChoiceHelp("unicode", "x", "display as hex <xx>")
	p.SetChoiceHelp("unicode", "h", "highlight escape in red")
	p.BindBool("o","radix-oct",&radix_o, "Same as -t o")
	p.BindBool("f","print-file-name",&print_file_name, "Print file name")
	p.BindBool("w","include-all-whitespace",&include_all_whitespace, "Include newlines/whitespace")
	p.BindStr("s", "output-separator", &custom_separator, false, "", "Custom separator")

	p.SetVersion("latest", "V", "strings unix command re-write in go")

	p.AddExclusive("t,o")
	p.AddExclusive("e,U")


	if err := p.Parse(os.Args[1:]); err != nil {
    if errors.Is(err, goargv.ErrHelpRequested) || errors.Is(err, goargv.ErrVersionRequested) {
        os.Exit(0)
    }

		switch {
		case errors.Is(err, goargv.ErrMutualExclusion):
			fmt.Fprintf(os.Stderr, "Conflict error: %v\n", err)
		case errors.Is(err, goargv.ErrDependency):
			fmt.Fprintf(os.Stderr, "Dependency error: %v\n", err)
		case errors.Is(err, goargv.ErrMissingRequired):
			fmt.Fprintf(os.Stderr, "Missing argument: %v\n", err)
		default:
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	if radix_o == true {
		radix = 'o'
	}
	switch encoding{
	case 's', 'S':
		byte_to_take = 1
	case 'b', 'l':
		byte_to_take = 2
	case 'B', 'L':
		byte_to_take = 4
	default:
		byte_to_take = 1
	}

	var printable_string strings.Builder
	var start_offset int

	for _, file := range source {
		fileBytes, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading bytes of file %s: %v.\n", file, err)
			continue
		}
		start_offset = 0
		var char_count = 0

		for i := 0; i < len(fileBytes); i += byte_to_take {
			var rune_block rune
			var block []byte
			if encoding != ' ' {
        if i+byte_to_take > len(fileBytes) {
            break
        }
				block = fileBytes[i : i+byte_to_take]
				rune_block = create_rune_from_block(block, encoding)
			}else{
				rune_block, byte_to_take = utf8.DecodeRune(fileBytes[i:])
				block = fileBytes[i : i+byte_to_take]	
			}

			if is_printable(rune_block, include_all_whitespace, unicode) {
				if char_count == 0{
					start_offset = i
				}
				char_count++
				if rune_block <= 127 {
					printable_string.WriteRune(rune_block)
				}else{

					if unicode == 'x' {
						for _, b := range block {
							printable_string.WriteString(fmt.Sprintf("<%02x>", b))
						}
					}else if unicode == 'e' {
						if rune_block <= 0xFFFF {
							printable_string.WriteString(fmt.Sprintf("\\u%04x", rune_block))
						} else{
							printable_string.WriteString(fmt.Sprintf("\\U%08x", rune_block))
						}
					}else if unicode == 'h' {
						var escape_seq string
						if rune_block <= 0xFFFF {
							escape_seq = fmt.Sprintf("\\u%04x", rune_block)
						}else{
							escape_seq = fmt.Sprintf("\\U%08x", rune_block)
						}
						printable_string.WriteString(fmt.Sprintf("\033[31;1m%s\033[0m", escape_seq))
					} else {
						printable_string.WriteRune(rune_block)
					}
				}
			}else{
				if char_count >= min_len {
					if print_file_name {
						fmt.Printf("%s: ",file)
					}
					if radix != ' ' {
						switch radix{
						case 'x':
							fmt.Printf("%7x ", start_offset)
						case 'd':
							fmt.Printf("%7d ", start_offset)
						case 'o':
							fmt.Printf("%7o ", start_offset)
						default:
							break
						}
					}

					fmt.Printf("%s%s", printable_string.String(), custom_separator)
				}
				printable_string.Reset()
				char_count = 0
			}
		}
		if char_count >= min_len {
			if print_file_name {
				fmt.Printf("%s: ", file)
			}
			if radix != ' ' {
				switch radix {
					case 'x': fmt.Printf("%7x ", start_offset)
					case 'd': fmt.Printf("%7d ", start_offset)
					case 'o': fmt.Printf("%7o ", start_offset)
				}
			}
			fmt.Printf("%s%s", printable_string.String(), custom_separator)
		}
		printable_string.Reset()
	}
}
