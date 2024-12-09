package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

const (
	getPasswordUrl = `http://pstgu.yss.su/iu9/networks/let1/getkey.php?hash=%s`
	sendEmailUrl   = `http://pstgu.yss.su/iu9/networks/let1_2024/send_from_go.php?subject=let1_ИУ9-32Б_Бойко_Роман&fio=Бойко_Роман&pass=%s`
)

func main() {
	cmd := exec.Command("tcpdump", "-l", "-A")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Ошибка создания pipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Ошибка запуска tcpdump:", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if containsHashStrong(line) {
			hash := extractHash(line)

			fmt.Println(line)

			if hash != "" {
				fmt.Println("Найден хэш пользователя:", hash)
			}

			pass, err := getPassword(hash)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(pass)

			err = sendEmail(pass)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else if containsHash(line) {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка чтения данных tcpdump:", err)
	}
}

func containsHash(line string) bool {
	return strings.Contains(line, "key:") && strings.Contains(line, "for")
}

func containsHashStrong(line string) bool {
	return strings.Contains(line, "key:") && strings.Contains(line, "for") && strings.Contains(line, "Bojko")
}

func extractHash(line string) string {
	words := bufio.NewScanner(strings.NewReader(line))
	words.Split(bufio.ScanWords)

	for words.Scan() {
		word := words.Text()
		if len(word) == 32 || len(word) == 40 {
			return word
		}
	}
	return ""
}

func getPassword(hash string) (string, error) {
	url := fmt.Sprintf(getPasswordUrl, hash)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	password := extractPassword(string(body))
	return password, nil
}

func extractPassword(body string) string {
	re := regexp.MustCompile(`pass:\s*([^\s]+)`)
	matches := re.FindStringSubmatch(body)

	if len(matches) > 1 {
		return matches[1] // Return the found password
	}

	return "" // Return an empty string if no password is found
}

func sendEmail(password string) error {
	url := fmt.Sprintf(sendEmailUrl, password)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending email: %s", resp.Status)
	}

	return nil
}
