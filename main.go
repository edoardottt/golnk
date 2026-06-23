package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Constants derived from official Microsoft documentation and LNK file analysis
const (
	headerSize             uint32 = 0x0000004C
	fileAttributeDirectory uint32 = 0x00000010
	fileAttributeArchive   uint32 = 0x00000020
	swShowNormal           uint32 = 0x00000001
)

var (
	// LinkCLSID: 00021401-0000-0000-c000-000000000046
	linkCLSID = []byte{0x01, 0x14, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}

	// CLSIDs for Computer and Network
	clsidComputer = []byte{0xe0, 0x4f, 0xd0, 0x20, 0xea, 0x3a, 0x69, 0x10, 0xa2, 0xd8, 0x08, 0x00, 0x2b, 0x30, 0x30, 0x9d} // My Computer
	clsidNetwork  = []byte{0x60, 0x2c, 0x8d, 0x20, 0xea, 0x3a, 0x69, 0x10, 0xa2, 0xd7, 0x08, 0x00, 0x2b, 0x30, 0x30, 0x9d} // Network Places

	// Prefixes found via LNK file analysis
	prefixLocalRoot    = []byte{0x2f}                                                                   // Local Disk
	prefixFolder       = []byte{0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // File Folder
	prefixFile         = []byte{0x32, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // File
	prefixNetworkRoot  = []byte{0xc3, 0x01, 0x81}                                                       // Network File Server Root
	prefixNetworkPrint = []byte{0xc3, 0x02, 0xc1}                                                       // Network Printer
)

// genDataString creates a string data block prefixed by its 16-bit size
func genDataString(s string) []byte {
	if s == "" {
		return nil
	}
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint16(len(s)))
	buf.WriteString(s)
	return buf.Bytes()
}

// genIDList prepends the 16-bit size (including the 2 bytes for the size itself) to the data
func genIDList(data []byte) []byte {
	size := uint16(len(data) + 2)
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, size)
	buf.Write(data)
	return buf.Bytes()
}

func main() {
	var (
		lnkTarget     string
		outputFile    string
		name          string
		workingDir    string
		arguments     string
		iconLocation  string
		isPrinterLink bool
	)

	// CLI Arguments setup
	flag.StringVar(&lnkTarget, "l", "", "Specifies the shortcut target")
	flag.StringVar(&lnkTarget, "lnk-target", "", "Specifies the shortcut target")
	flag.StringVar(&outputFile, "o", "", "Saves the shortcut to a file")
	flag.StringVar(&outputFile, "output-file", "", "Saves the shortcut to a file")
	flag.StringVar(&name, "n", "", "Specifies a description for the shortcut")
	flag.StringVar(&name, "name", "", "Specifies a description for the shortcut")
	flag.StringVar(&workingDir, "w", "", "Specifies the working directory for the command")
	flag.StringVar(&workingDir, "working-dir", "", "Specifies the working directory for the command")
	flag.StringVar(&arguments, "a", "", "Specifies the arguments for the command")
	flag.StringVar(&arguments, "arguments", "", "Specifies the arguments for the command")
	flag.StringVar(&iconLocation, "i", "", "Specifies the icon path")
	flag.StringVar(&iconLocation, "icon", "", "Specifies the icon path")
	flag.BoolVar(&isPrinterLink, "p", false, "Generates a network printer shortcut")
	flag.BoolVar(&isPrinterLink, "printer-link", false, "Generates a network printer shortcut")

	// Custom Usage message to match the bash script
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage:\n %s -l shortcut_target [-n description] [-w working_dir] [-a cmd_args] [-i icon_path] -o my_file.lnk [-p]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, " -l, --lnk-target               Specifies the shortcut target\n")
		fmt.Fprintf(os.Stderr, " -o, --output-file              Saves the shortcut to a file\n")
		fmt.Fprintf(os.Stderr, " -n, --name                     Specifies a description for the shortcut\n")
		fmt.Fprintf(os.Stderr, " -w, --working-dir              Specifies the working directory for the command\n")
		fmt.Fprintf(os.Stderr, " -a, --arguments                Specifies the arguments for the command\n")
		fmt.Fprintf(os.Stderr, " -i, --icon                     Specifies the icon path\n")
		fmt.Fprintf(os.Stderr, " -p, --printer-link             Generates a network printer shortcut\n\n")
	}

	flag.Parse()

	// Handle unknown options left in the arguments list
	if len(flag.Args()) > 0 {
		fmt.Fprintf(os.Stderr, "Unknown option(s): %v\n", flag.Args())
		os.Exit(1)
	}

	if lnkTarget == "" || outputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Calculate LinkFlags and String Data
	var stringDataBuf bytes.Buffer
	linkFlags := uint32(0x0101) // HasLinkTargetIDList (0x01) + ForceNoLinkInfo (0x0100)

	if name != "" {
		linkFlags |= 0x04 // HasName
		stringDataBuf.Write(genDataString(name))
	}
	if workingDir != "" {
		linkFlags |= 0x10 // HasWorkingDir
		stringDataBuf.Write(genDataString(workingDir))
	}
	if arguments != "" {
		linkFlags |= 0x20 // HasArguments
		stringDataBuf.Write(genDataString(arguments))
	}
	if iconLocation != "" {
		linkFlags |= 0x40 // HasIconLocation
		stringDataBuf.Write(genDataString(iconLocation))
	}

	// Remove trailing backslash if present
	lnkTarget = strings.TrimRight(lnkTarget, "\\")

	var isNetworkLnk bool
	var isRootLnk bool
	var targetRoot string
	var targetLeaf string
	var prefixRoot []byte
	var itemData []byte

	// Separate root path from final target
	if strings.HasPrefix(lnkTarget, `\\`) {
		isNetworkLnk = true
		prefixRoot = prefixNetworkRoot
		itemData = append([]byte{0x1f, 0x58}, clsidNetwork...)

		lastSlashIndex := strings.LastIndex(lnkTarget, `\`)
		if lastSlashIndex > 1 {
			targetRoot = lnkTarget[:lastSlashIndex]
			targetLeaf = lnkTarget[lastSlashIndex+1:]
		} else {
			targetRoot = lnkTarget
		}
	} else {
		prefixRoot = prefixLocalRoot
		itemData = append([]byte{0x1f, 0x50}, clsidComputer...)

		firstSlashIndex := strings.Index(lnkTarget, `\`)
		if firstSlashIndex != -1 {
			targetRoot = lnkTarget[:firstSlashIndex]
			lastSlashIndex := strings.LastIndex(lnkTarget, `\`)
			if lastSlashIndex != -1 && lastSlashIndex >= firstSlashIndex {
				targetLeaf = lnkTarget[lastSlashIndex+1:]
			}
		} else {
			targetRoot = lnkTarget
		}

		if !strings.HasSuffix(targetRoot, `\`) {
			targetRoot += `\`
		}
	}

	if isPrinterLink {
		prefixRoot = prefixNetworkPrint
		targetRoot = lnkTarget
		isRootLnk = true
	}

	if len(targetLeaf) == 0 {
		isRootLnk = true
	}

	// Select prefix for the target to display the right icon
	var prefixOfTarget []byte
	var fileAttributes uint32
	typeTarget := "folder"

	// Match 3-character extensions (like *.??? in bash)
	if len(targetLeaf) >= 4 {
		dotIdx := strings.LastIndex(targetLeaf, ".")
		if dotIdx != -1 && len(targetLeaf)-dotIdx == 4 {
			prefixOfTarget = prefixFile
			typeTarget = "file"
			fileAttributes = fileAttributeArchive
		}
	}

	if typeTarget == "folder" {
		prefixOfTarget = prefixFolder
		fileAttributes = fileAttributeDirectory
	}

	// Target Root needs 21 null bytes appended (Required from Vista onwards)
	targetRootBytes := []byte(targetRoot)
	targetRootBytes = append(targetRootBytes, make([]byte, 21)...)

	// Build IDLIST items
	var idListItems []byte
	idListItems = append(idListItems, genIDList(itemData)...)

	rootPayload := append(prefixRoot, targetRootBytes...)
	rootPayload = append(rootPayload, 0x00) // End of string
	idListItems = append(idListItems, genIDList(rootPayload)...)

	if !isRootLnk {
		leafPayload := append(prefixOfTarget, []byte(targetLeaf)...)
		leafPayload = append(leafPayload, 0x00) // End of string
		idListItems = append(idListItems, genIDList(leafPayload)...)
	}

	// Determine Link Type for console output
	typeLnk := "local"
	if isNetworkLnk {
		typeLnk = "network"
		if isPrinterLink {
			typeTarget = "printer"
		}
	}

	fmt.Fprintf(os.Stderr, "Creating a shortcut of type \"%s %s\" targeting %s %s\n", typeTarget, typeLnk, lnkTarget, arguments)

	// Assemble final binary file
	buf := new(bytes.Buffer)

	_ = binary.Write(buf, binary.LittleEndian, headerSize)
	buf.Write(linkCLSID)
	_ = binary.Write(buf, binary.LittleEndian, linkFlags)
	_ = binary.Write(buf, binary.LittleEndian, fileAttributes)
	buf.Write(make([]byte, 24))                              // CreationTime, AccessTime, WriteTime (8 bytes each)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))    // FileSize
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))    // IconIndex
	_ = binary.Write(buf, binary.LittleEndian, swShowNormal) // ShowCommand
	buf.Write(make([]byte, 12))                              // Hotkey (2), Reserved (2), Reserved2 (4), Reserved3 (4)

	buf.Write(genIDList(idListItems)) // IDLIST
	buf.Write([]byte{0x00, 0x00})     // TerminalID

	buf.Write(stringDataBuf.Bytes()) // STRING_DATA

	// Write buffer to file
	err := os.WriteFile(outputFile, buf.Bytes(), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}
}
