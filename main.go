package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/projectdiscovery/goflags"
)

var urls, paths, methods, backFileName []string

var options *Options
var extensions = []string{"", ".7z", ".bakoms.zip", ".db", ".gz", ".iml", ".jar", ".log", ".mdb", ".pem", ".rar", ".zip", ".tar.gz", ".tar", ".bz2", ".sql", ".backup", ".war", ".bak", ".dll", ".root ", ".sql.gz", ".tar.bz2", ".tar.tgz", ".tgz", ".txt"}

func main() {

	options = ParseOptions()

	if !options.silent {
		timeInfo("starting")
		defer timeInfo("ending")
		banner()
	}

	if options.file != "" {
		readFromFile()
	} else {
		if options.url == "" {
			fmt.Println("[!] 请输入url!")
			os.Exit(1)
		} else {
			readFromStdin()
		}
	}

	if options.extension != "" {
		extensions = strings.Split(options.extension, ",")
	}

	if options.method == "all" {
		m := "regular,withoutdots,withoutvowels,reverse,mixed,withoutdv,shuffle"
		methods = strings.Split(m, ",")
	} else {
		methods = strings.Split(options.method, ",")
	}

	if options.paths != "/" {
		paths = strings.Split(options.paths, ",")
	} else {
		paths = strings.Split(options.paths, "")
	}

	wp := workerpool.New(options.worker)

	for _, url := range urls {
		url := url
		wp.Submit(func() {
			start(url)
		})
	}
	wp.StopWait()
}

func start(domain string) {
	var rgx = regexp.MustCompile(options.exclude)
	if len(domain) < options.domain_length+8 {
		if !rgx.MatchString(domain) {
			getAllCombination(domain)
		}
	}
}

func getAllCombination(domain string) string {
	var generateWordlist []string //这个变量会存储所有的切片值
	var newWordlist []string

	for _, method := range methods {
		switch method {
		case "regular":
			regularDomain(domain, &generateWordlist)
		case "withoutdots":
			withoutDots(domain, &generateWordlist)
		case "withoutvowels":
			withoutVowels(domain, &generateWordlist)
		case "reverse":
			reverseDomain(domain, &generateWordlist)
		case "mixed":
			mixedSubdomain(domain, &generateWordlist)
		case "withoutdv":
			withoutVowelsAndDots(domain, &generateWordlist)
		case "shuffle":
			shuffle(domain, &generateWordlist)
		default:
			shuffle(domain, &generateWordlist)
		}
	}

	if options.just_wordlist {
		for _, word := range generateWordlist {
			for _, e := range extensions {
				for _, path := range paths {
					url := domain + path + options.prefix + word + options.suffix + e
					fmt.Println(url)
				}
			}
		}

		return ""
	}
	for _, tmp := range generateWordlist {
		if tmp[len(tmp)-1] == '.' {
			tmp = tmp[:len(tmp)-1]
		}
		newWordlist = append(newWordlist, tmp)
	}
	joinExtensions(RemoveDuplicationMap(newWordlist))

	currentTime := time.Now()
	timestamp := currentTime.Format("20060102150405")
	file, err := os.Create(fmt.Sprintf("%s_out.txt", timestamp))
	if err != nil {
		log.Fatal("[-]", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)

	for _, word := range backFileName {
		_, err := writer.WriteString(word + "\n")
		if err != nil {
			fmt.Printf("[-]%s\n", err)
		}
		//wpx.Submit(func() {
		//	headRequest(domain, word)
		//})

	}
	return ""
}

// RemoveDuplicationMap 数据去重
func RemoveDuplicationMap(arr []string) []string {
	set := make(map[string]struct{}, len(arr))
	j := 0
	for _, v := range arr {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		arr[j] = v
		j++
	}

	return arr[:j]
}

func joinExtensions(generate_wordlist []string) {
	for _, word := range generate_wordlist {
		for _, e := range extensions {

			backFileName = append(backFileName, fmt.Sprintf("%s%s", word, e))
		}
	}
}

func regularDomain(domain string, wordlist *[]string) {
	generatePossibilities(domain, wordlist)
}

func withoutDots(domain string, wordlist *[]string) {
	without_dot := strings.ReplaceAll(domain, ".", "")
	generatePossibilities(without_dot, wordlist)
}

func withoutVowels(domain string, wordlist *[]string) {
	clear_vowel := strings.NewReplacer("a", "", "e", "", "i", "", "u", "", "o", "")
	domain_without_vowel := clear_vowel.Replace(domain)
	generatePossibilities(domain_without_vowel, wordlist)
}

func withoutVowelsAndDots(domain string, wordlist *[]string) {
	clear_vowel := strings.NewReplacer("a", "", "e", "", "i", "", "u", "", "o", "", ".", "")
	without_vowel_dot := clear_vowel.Replace(domain)
	generatePossibilities(without_vowel_dot, wordlist)
}

func mixedSubdomain(domain string, wordlist *[]string) {
	clear_domain := strings.Split(domain, "://")[1]
	split := strings.Split(clear_domain, ".")

	for sindex := range split {
		for eindex := range split {
			generatePossibilities("http://"+split[sindex]+"."+split[eindex], wordlist)
		}
	}
}

func reverseDomain(domain string, wordlist *[]string) {
	clear_domain := strings.Split(domain, "://")[1]
	split := strings.Split(clear_domain, ".")
	split_reverse := reverseSlice(split)
	reverse_domain := "http://" + strings.Join(split_reverse, ".")
	generatePossibilities(reverse_domain, wordlist)
	withoutDots(reverse_domain, wordlist)
	withoutVowels(reverse_domain, wordlist)
	withoutVowelsAndDots(reverse_domain, wordlist)
}

func shuffle(domain string, wordlist *[]string) {
	clear_domain := strings.Split(domain, "://")[1]
	split := strings.Split(clear_domain, ".")
	split_reverse := reverseSlice(split)
	reverse_domain := "http://" + strings.Join(split_reverse, ".")
	shuffleSubdomain(domain, wordlist)
	shuffleSubdomain(reverse_domain, wordlist)
}

func shuffleSubdomain(domain string, wordlist *[]string) {
	clear_domain := strings.Split(domain, "://")[1]
	splt := strings.Split(clear_domain, ".")
	for id1 := range splt {
		for id2 := range splt[id1:] {
			p := strings.Join(splt[id1:id1+id2+1], ".")
			addShuffleSubdomain(p, wordlist)
			if id2 >= 2 {
				p = splt[id1] + "." + splt[id1+id2]
				addShuffleSubdomain(p, wordlist)
			}
		}
	}
}

func addShuffleSubdomain(domain string, wordlist *[]string) {
	if !contains(*wordlist, domain) {
		*wordlist = append(*wordlist, domain)
	}

	clear_vowel := strings.NewReplacer("a", "", "e", "", "i", "", "u", "", "o", "")
	domain_without_vowel := clear_vowel.Replace(domain)
	if !contains(*wordlist, domain_without_vowel) {
		*wordlist = append(*wordlist, domain_without_vowel)
	}

	without_dot := strings.ReplaceAll(domain, ".", "")
	if !contains(*wordlist, without_dot) {
		*wordlist = append(*wordlist, without_dot)
	}

	clear_voweldot := strings.NewReplacer("a", "", "e", "", "i", "", "u", "", "o", "", ".", "")
	without_vowel_dot := clear_voweldot.Replace(domain)
	if !contains(*wordlist, without_vowel_dot) {
		*wordlist = append(*wordlist, without_vowel_dot)
	}
}

func contains(slice []string, elements string) bool {
	for _, s := range slice {
		if elements == s {
			return true
		}
	}
	return false
}

func reverseSlice(slice []string) []string {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

func generatePossibilities(domain string, possibilities *[]string) {
	just_domain := strings.Split(domain, "://")[1]
	for first := range just_domain {
		for last := range just_domain[first:] {
			p := just_domain[first : first+last+1]
			if !contains(*possibilities, p) {
				*possibilities = append(*possibilities, p)
			}
		}
	}
}

func readFromStdin() {
	//scanner := bufio.NewScanner(os.Stdin)
	//scanner.Split(bufio.ScanLines)
	//
	//for scanner.Scan() {
	//	urls = append(urls, scanner.Text())
	//}
	// 将管道删除,从flag中接收值,此处为不破坏结构,转化为切片格式
	urls = append(urls, options.url)
}

func readFromFile() {
	file, err := os.Open(options.file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
}

func timeInfo(t string) {
	ctime := fmt.Sprintf("\n[*] Scan "+t+" time: %s", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println(("\033[36m") + ctime + ("\033[0m"))
}

type Options struct {
	contentType        string
	httpMethod         string
	userAgent          string
	extension          string
	exclude            string
	replace            string
	method             string
	prefix             string
	suffix             string
	remove             string
	paths              string
	url                string
	file               string
	proxy              string
	worker             int
	timeout            int
	status_code        int
	domain_length      int
	min_content_length int
	just_wordlist      bool
	silent             bool
	version            bool
	print              bool
	help               bool
}

func ParseOptions() *Options {
	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`fuzzuli is a url fuzzing tool that aims to find critical backup files by creating a dynamic wordlist based on the domain.`)

	createGroup(flagSet, "General Options", "GENERAL OPTIONS",
		flagSet.StringVar(&options.url, "u", "", "input host/domain"),
		flagSet.BoolVar(&options.version, "v", false, "print version"),
	)

	createGroup(flagSet, "wordlist options", "WORDLIST OPTIONS",
		flagSet.StringVar(&options.method, "mt", "", "methods. avaible methods: regular, withoutdots, withoutvowels, reverse, mixed, withoutdv, shuffle"),
		flagSet.StringVar(&options.suffix, "sf", "", "suffix"),
		flagSet.StringVar(&options.prefix, "pf", "", "prefix"),
		flagSet.StringVar(&options.extension, "ex", "", "file extension. default (.7z, .bakoms.zip, .db, .gz, .iml, .jar, .log, .mdb, .pem, .rar, .zip, .tar.gz, .tar, .bz2, .sql, .backup, .war, .bak, .dll, .root , .sql.gz, .tar.bz2, .tar.tgz, .tgz, .txt)"),
		flagSet.StringVar(&options.replace, "rp", "", "replace specified char"),
		flagSet.StringVar(&options.remove, "rm", "", "remove specified char"),
		flagSet.BoolVar(&options.just_wordlist, "jw", false, "just generate wordlist do not http request"),
	)

	createGroup(flagSet, "domain options", "DOMAIN OPTIONS",
		flagSet.StringVar(&options.exclude, "es", "#", "exclude domain that contains specified string or char. e.g. for OR operand google|bing|yahoo"),
		flagSet.IntVar(&options.domain_length, "dl", 40, "match domain length that specified."),
	)

	_ = flagSet.Parse()

	Version := "v0.0.1"
	if options.version {
		fmt.Println("Current Version:", Version)
		os.Exit(0)
	}

	return options
}

func createGroup(flagSet *goflags.FlagSet, groupName, description string, flags ...*goflags.FlagData) {
	flagSet.SetGroup(groupName, description)
	for _, currentFlag := range flags {
		currentFlag.Group(groupName)
	}
}

func banner() {
	fmt.Println(`

                _        ______             _____  _      _   
     /\        | |      |  ____|           |  __ \(_)    | |  
    /  \  _   _| |_ ___ | |__ _   _ _______| |  | |_  ___| |_ 
   / /\ \| | | | __/ _ \|  __| | | |_  /_  / |  | | |/ __| __|
  / ____ \ |_| | || (_) | |  | |_| |/ / / /| |__| | | (__| |_ 
 /_/    \_\__,_|\__\___/|_|   \__,_/___/___|_____/|_|\___|\__|

Base @musana | xbfding
--------------------------------------------`)

}
