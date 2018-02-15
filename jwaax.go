package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var _secret = ""
var _domain = ""
var _indrak = ""

func getSecret() string {
	if _secret != "" {
		return _secret
	}
	dat, err := ioutil.ReadFile("site_secret.txt")
	if err != nil {
		panic(err)
	}
	key := string(dat)

	_domain = key[0:strings.Index(key, ":")]
	_secret = key[strings.Index(key, ":")+1 : len(key)-1]
	//println("DOMAIN: " + _domain)
	//println("SECRET: " + _secret)
	return _secret
}

func GetIndraKey() string {
	if _indrak != "" {
		return _indrak
	}

	resp, err := http.Get("https://secure.demilletech.net/api/key")
	if err != nil {
		fmt.Println("## Indra is down! PANIC! PANIC! ##")
		return "Wha? Indra is down? Nuuuu!"
	}
	defer resp.Body.Close()
	keyb, _ := ioutil.ReadAll(resp.Body)
	key := string(keyb)

	_indrak = key
	return _indrak
}

func GetDomain() string {
	getSecret()
	return _domain
}

func GetEpochTime() int64 {
	now := time.Now()
	secs := now.Unix()
	return secs // hehe
}

func VerifyToken(token string, keya ...string) bool {
	key := getSecret()
	if len(keya) > 0 {
		key = keya[0]
	}

	ret := DecodeToken(token, key)
	if resp, ok := ret["RESPONSE"]; ok {
		if resp == "ERROR" {
			for key, value := range ret {
				println(key + " : " + value)
			}
			return false
		}
	}
	return true
}

func encodeToken(payload map[string]interface{}) string {

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
	tokenString, err := token.SignedString([]byte(getSecret()))
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		return "ERROR"
	}

	return tokenString
}

func GenerateToken(uniqueid string, returl string) string {
	payload := jwt.MapClaims{
		"aud":      "secure.demilletech.net",
		"domain":   GetDomain(),
		"uniqueid": uniqueid,
		"returl":   returl,
		"iat":      strconv.Itoa(int(GetEpochTime())),
		"exp":      strconv.Itoa(int(GetEpochTime() + 120)),
		"iss":      GetDomain(),
	}
	return encodeToken(payload)
}

func DecodeToken(encodedToken string, key string) map[string]string {

	if key == "#INDRAK#" {
		key = GetIndraKey()
	}

	token, err := jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:

		if strings.Contains(key, "BEGIN PUBLIC KEY") {
			if _, ok := token.Method.(*jwt.SigningMethodRSAPSS); ok {
				cleankey, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(key))
				return cleankey, nil
			} else {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		} else {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); ok {
				return key, nil
			} else {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
		}

		// This should never be used, but it's a safety
		return key, nil
	})

	if !token.Valid { // Meh, kinda faster maybe because I don't 100% trust the compiler?
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				println("INVALID_TOKEN")
				return map[string]string{
					"RESPONSE": "ERROR",
					"MESSAGE":  "INVALID_TOKEN",
				}
			} else if ve.Errors&(jwt.ValidationErrorNotValidYet) != 0 {
				println("NOT_VALID_YET")
				return map[string]string{
					"RESPONSE": "ERROR",
					"MESSAGE":  "NOT_VALID_YET",
				}
			} else if ve.Errors&(jwt.ValidationErrorExpired) != 0 {
				println("EXPIRED_TOKEN")
				return map[string]string{
					"RESPONSE": "ERROR",
					"MESSAGE":  "EXPIRED_TOKEN",
				}
			} else if ve.Errors&(jwt.ValidationErrorSignatureInvalid) != 0 {
				println("INVALID_SIGNATURE")
				return map[string]string{
					"RESPONSE": "ERROR",
					"MESSAGE":  "INVALID_SIGNATURE",
				}
			} else if ve.Errors&(jwt.ValidationErrorUnverifiable) != 0 {
				println("UNVERIFIABLE_TOKEN")
				return map[string]string{
					"RESPONSE": "ERROR",
					"MESSAGE":  "UNVERIFIABLE_TOKEN",
				}
			} else if ve.Errors&(jwt.ValidationErrorAudience) != 0 {
				println("INVALID_AUDIENCE")
				return map[string]string{
					"RESPONSE": "ERROR",
					"MESSAGE":  "INVALID_AUDIENCE",
				}
			}

			println("NO_CLUE")
			return map[string]string{
				"RESPONSE": "ERROR",
				"MESSAGE":  "NO_CLUE",
			}
		}
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ret := map[string]string{}
		for key, value := range claims {
			ret[key] = value.(string)
		}

		return ret
	}
	return map[string]string{"RESPONSE": "ERROR"}
}
