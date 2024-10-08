package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"signalone/cmd/config"
	"signalone/pkg/models"
	"unicode"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func CalculateNewCounter(currentScore int32, newScore int32, counter int32) int32 {
	return counter + (newScore - currentScore)
}

func GenerateFilter(fields bson.M, operator string) bson.M {
	conditions := make([]bson.M, 0, len(fields))

	for field, value := range fields {
		conditions = append(conditions, bson.M{field: value})
	}

	return bson.M{operator: conditions}
}

func CallPredictionAgentService(jsonData []byte) (analysisResponse models.IssueAnalysis, err error) {
	var cfg = config.GetInstance()

	issueAnalysisReq, err := http.NewRequest("POST", cfg.PredicitonAgentServiceUrl+"/run_analysis", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	issueAnalysisReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(issueAnalysisReq)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("error: %s", resp.Status)
		return
	}
	defer resp.Body.Close()
	rawAnalysisResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(rawAnalysisResponse, &analysisResponse)
	if err != nil {
		return
	}
	return
}

func CallCodeGenAgentService(jsonData []byte) (analysisResponse models.CodeSnippetResponse, err error) {
	var cfg = config.GetInstance()

	issueAnalysisReq, err := http.NewRequest("POST", cfg.PredicitonAgentServiceUrl+"/generate_code_snippet", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	issueAnalysisReq.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(issueAnalysisReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	rawAnalysisResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(rawAnalysisResponse, &analysisResponse)
	if err != nil {
		return
	}
	return
}

func CompareLogs(incomingLogTails []string, currentIssuesLogTails []string) (isNewIssue bool) {
	const LogSimilarityThreshold = 0.644

	isNewIssue = true
	sdm := metrics.NewSorensenDice()
	sdm.CaseSensitive = false
	sdm.NgramSize = 3
	for _, incomingLogTail := range incomingLogTails {
		for _, currentIssueLogTail := range currentIssuesLogTails {
			similarity := strutil.Similarity(incomingLogTail, currentIssueLogTail, sdm)
			if similarity >= LogSimilarityThreshold {
				isNewIssue = false
				return
			}
		}
	}
	return
}

func FilterForRelevantLogs(logs []string) []string {
	var relevantLogs = make([]string, 0)
	//Classes are absractions of different types of logs as different types of issues
	// Class 0 = Error messages
	// Class 1 = Warning messages
	// Class 2 = Info messages
	issueClassZeroRegex := `(?i)(abort|blocked|corrupt|crash|critical|deadlock|
		denied|err|error|exception|fatal|forbidden|
		freeze|hang|illegal|invalid|missing|panic|refused|rejected|stacktrace|
		timeout|traceback|unauthorized|uncaught|undefined|unhandled|unsupported)`
	issueClassOneregex := `(?i)(deprecated|deprecating|warn|warning)`

	compiledClassZeroRegex := regexp.MustCompile(issueClassZeroRegex)
	compiledClassOneRegex := regexp.MustCompile(issueClassOneregex)

	globalLoopMatched := false

	globalLoopMatched = executeRelevantLogsLoopComponent(logs, &relevantLogs, compiledClassZeroRegex)
	if !globalLoopMatched {
		executeRelevantLogsLoopComponent(logs, &relevantLogs, compiledClassOneRegex)
	}

	return relevantLogs
}

func ComparePasswordHashes(hashedPassword string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func PasswordValidation(password string) bool {
	if !(len(password) >= 8 && len(password) <= 50) {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		}
		if unicode.IsLower(char) {
			hasLower = true
		}
		if unicode.IsDigit(char) {
			hasDigit = true
		}
	}
	return (hasUpper && hasLower && hasDigit)
}

func AnonymizePII(data string) string {
	const EmailRegex = `(?i)([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`
	const PhoneRegex = `(?i)(\+?(\d{1,3})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4})`
	const CreditCardRegex = `(?i)(\b(?:\d[ -]*?){13,16}\b)`

	compiledEmailRegex := regexp.MustCompile(EmailRegex)
	compiledPhoneRegex := regexp.MustCompile(PhoneRegex)
	compiledCreditCardRegex := regexp.MustCompile(CreditCardRegex)

	data = compiledEmailRegex.ReplaceAllString(data, "[email]")
	data = compiledPhoneRegex.ReplaceAllString(data, "[phone]")
	data = compiledCreditCardRegex.ReplaceAllString(data, "[credit card]")

	return data
}

func MaskSecrets(data string) string {
	const PasswordRegex = `(?i)(\bpassword\s*[=:\s]\s*[A-Za-z0-9]{8,}\b)`
	const BearerTokenRegex = `(?i)(\bBearer\s[A-Za-z0-9_#@&*=]{8,}\b)`
	const BasicAuthRegex = `(?i)(\bBasic\s[A-Za-z0-9]{8,}\b)`
	const JWTRegex = `(?i)(\b[A-Za-z0-9_#@&*=]{8,}\.[A-Za-z0-9_#@&*=]{8,}\.[A-Za-z0-9_#@&*=]{8,}\b)`

	compiledPasswordRegex := regexp.MustCompile(PasswordRegex)
	compiledBearerTokenRegex := regexp.MustCompile(BearerTokenRegex)
	compiledBasicAuthRegex := regexp.MustCompile(BasicAuthRegex)
	compiledJWTRegex := regexp.MustCompile(JWTRegex)

	data = compiledPasswordRegex.ReplaceAllString(data, "[REDUCTED]")
	data = compiledBearerTokenRegex.ReplaceAllString(data, "[REDUCTED]")
	data = compiledBasicAuthRegex.ReplaceAllString(data, "[REDUCTED]")
	data = compiledJWTRegex.ReplaceAllString(data, "[REDUCTED]")

	return data
}
func executeRelevantLogsLoopComponent(logs []string, relevantLogs *[]string, regEx *regexp.Regexp) bool {
	var globalLoopMatched = false
	for logIndex, log := range logs {
		if len(*relevantLogs) != 0 {
			for _, relevantLog := range *relevantLogs {
				if log == relevantLog {
					continue
				}
			}
		}
		if matched := regEx.MatchString(log); matched {
			*relevantLogs = append(*relevantLogs, logs[logIndex])
			globalLoopMatched = true
			//Add the next and previous log to the relevant logs if stack trace is found
			//To be improved
			if logIndex+1 < len(logs) {
				*relevantLogs = append(*relevantLogs, logs[logIndex+1])
			}
			if logIndex-1 >= 0 {
				*relevantLogs = append(*relevantLogs, logs[logIndex-1])
			}
		}
	}
	return globalLoopMatched
}
