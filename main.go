package main

import (
	"fmt"
	"index/suffixarray"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

func filename() (string, error) {
	var filename string
	if filename = os.Getenv("AWS_SHARED_CREDENTIALS_FILE"); filename != "" {
		return filename, nil
	}

	homeDir := os.Getenv("HOME") // *nix
	if homeDir == "" {           // Windows
		homeDir = os.Getenv("USERPROFILE")
	}

	if homeDir == "" {
		return "", fmt.Errorf("user home directory not found")
	}

	filename = filepath.Join(homeDir, ".aws", "credentials")

	return filename, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("expected profile")
	}

	profileName := os.Args[1]

	filename, err := filename()
	if err != nil {
		log.Fatal(err)
	}
	config, err := ini.Load(filename)
	if err != nil {
		log.Fatal(err)
	}

	if profileName == "default" {
		return
	}

	sectionName, err := findSection(config.SectionStrings(), profileName)
	if err != nil {
		log.Fatal(err)
	}

	profile, err := config.GetSection(sectionName)
	if err != nil {
		log.Fatalf("profile %q not found", profileName)
	}
	var archiveProfile *ini.Section
	defaultProfile, _ := config.GetSection("default")
	for _, section := range config.Sections() {
		if sectionEqual(defaultProfile, section) {
			archiveProfile = section
		}
	}
	if archiveProfile == nil {
		archiveProfile, err := config.NewSection("archive_" + time.Now().Format("20060102150405"))
		if err != nil {
			log.Fatal("could not create archive profile")
		}
		for _, key := range defaultProfile.Keys() {
			_, err := archiveProfile.NewKey(key.Name(), key.Value())
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	config.DeleteSection("default")
	newDefaultProfile, err := config.NewSection("default")
	if err != nil {
		log.Fatal("could not create default profile")
	}
	for _, key := range profile.Keys() {
		_, err := newDefaultProfile.NewKey(key.Name(), key.Value())
		if err != nil {
			log.Fatal(err)
		}
	}

	err = config.SaveTo(filename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Switched aws profile %q to default.", sectionName)
}

func findSection(sections []string, profileName string) (string, error) {
	joinedStrings := "\x00" + strings.Join(sections, "\x00")
	sa := suffixarray.New([]byte(joinedStrings))

	// User has typed in "he"
	match, err := regexp.Compile("\x00" + profileName + "[^\x00]*")
	if err != nil {
		return "", err
	}
	ms := sa.FindAllIndex(match, -1)
	if len(ms) == 0 {
		return "", fmt.Errorf("no matching profile found")
	}
	if len(ms) > 1 {
		profiles := make([]string, len(ms))
		for i, m := range ms {
			profiles[i] = joinedStrings[m[0]+1 : m[1]]
		}
		return "", fmt.Errorf("multiple profiles found: %v", profiles)
	}
	return joinedStrings[ms[0][0]+1 : ms[0][1]], nil
}

func sectionEqual(s1, s2 *ini.Section) bool {
	if len(s1.Keys()) != len(s2.Keys()) {
		return false
	}
	for _, key := range s1.Keys() {
		if key2, err := s2.GetKey(key.Name()); err != nil || key2.Value() != key.Value() {
			return false
		}
	}
	return true
}
