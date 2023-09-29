package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

var userSecret string

func checkPort(ip string, port int, wg *sync.WaitGroup) {
	defer wg.Done()

	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)

	if err == nil {
		fmt.Printf("Port %d is open\n", port)
		conn.Close()

		apiUrlSignup := fmt.Sprintf("http://%s:%d/signup", ip, port)
		userName := "secretme"
		requestDataSignup := []byte(`{"User": "` + userName + `"}`)

		respSignup, err := http.Post(apiUrlSignup, "application/json", bytes.NewBuffer(requestDataSignup))
		if err != nil {
			fmt.Println("Erreur lors de l'envoi de la demande POST à l'API pour /signup:", err)
			return
		}
		defer respSignup.Body.Close()

		if respSignup.Status == "200 OK" {
			fmt.Printf("Inscription réussie pour l'utilisateur : %s\n", userName)
		} else {
			fmt.Printf("Échec de l'inscription pour l'utilisateur : %s\n", userName)
		}

		apiUrlCheck := fmt.Sprintf("http://%s:%d/check", ip, port)
		requestDataCheck := []byte(`{"User": "` + userName + `"}`)

		respCheck, err := http.Post(apiUrlCheck, "application/json", bytes.NewBuffer(requestDataCheck))
		if err != nil {
			fmt.Println("Erreur lors de l'envoi de la demande POST à l'API pour /check:", err)
			return
		}
		defer respCheck.Body.Close()

		if respCheck.Status == "200 OK" {
			fmt.Printf("Vérification réussie pour l'utilisateur : %s\n", userName)
		} else {
			fmt.Printf("Échec de la vérification pour l'utilisateur : %s\n", userName)
		}

		apiUrlPing := fmt.Sprintf("http://%s:%d/ping", ip, port)

		respPing, err := http.Get(apiUrlPing)
		if err != nil {
			fmt.Println("Erreur lors de l'envoi de la demande GET à l'API pour /ping:", err)
			return
		}
		defer respPing.Body.Close()

		if respPing.Status == "200 OK" {
			fmt.Printf("Ping réussi pour le port : %d\n", port)
		} else {
			fmt.Printf("Échec du ping pour le port : %d\n", port)
		}

		apiUrlGetUserSecret := fmt.Sprintf("http://%s:%d/getUserSecret", ip, port)
		requestDataGetUserSecret := []byte(`{"User": "` + userName + `"}`)

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		for {
			respGetUserSecret, err := client.Post(apiUrlGetUserSecret, "application/json", bytes.NewBuffer(requestDataGetUserSecret))
			if err != nil {
				fmt.Println("Erreur lors de l'envoi de la demande POST à l'API pour /getUserSecret:", err)

				
				if isConnectionError(err) {
					fmt.Println("La connexion a été réinitialisée, réessai en cours...")
					time.Sleep(5 * time.Second)
					continue
				}

				return
			}
			defer respGetUserSecret.Body.Close()

			if respGetUserSecret.Status == "200 OK" {
				secretBody := make([]byte, 64)
				_, err := respGetUserSecret.Body.Read(secretBody)
				if err != nil {
					fmt.Println("Erreur lors de la lecture du secret:", err)
					return
				}
				userSecret = string(secretBody)
				fmt.Printf("Récupération du secret réussie pour l'utilisateur : %s\n", userName)
				break 
			} else if respGetUserSecret.Status == "500 Internal Server Error" {
				fmt.Println("Le serveur n'est pas prêt, réessai en cours...")
				time.Sleep(5 * time.Second)
			} else {
				fmt.Printf("Échec de la récupération du secret pour l'utilisateur : %s\n", userName)
				break 
			}
		}

		apiUrlGetUserLevel := fmt.Sprintf("http://%s:%d/getUserLevel", ip, port)
		requestDataGetUserLevel := []byte(`{"User": "` + userName + `", "Secret": "` + userSecret + `"}`)

		respGetUserLevel, err := http.Post(apiUrlGetUserLevel, "application/json", bytes.NewBuffer(requestDataGetUserLevel))
		if err != nil {
			fmt.Println("Erreur lors de l'envoi de la demande POST à l'API pour /getUserLevel:", err)
			return
		}
		defer respGetUserLevel.Body.Close()

		if respGetUserLevel.Status == "200 OK" {
			fmt.Printf("Niveau de l'utilisateur : %s\n", userName)
		} else {
			fmt.Printf("Échec de la récupération du niveau de l'utilisateur : %s\n", userName)
		}

		apiUrlGetUserPoints := fmt.Sprintf("http://%s:%d/getUserPoints", ip, port)
		requestDataGetUserPoints := []byte(`{"User": "` + userName + `", "Secret": "` + userSecret + `"}`)

		respGetUserPoints, err := http.Post(apiUrlGetUserPoints, "application/json", bytes.NewBuffer(requestDataGetUserPoints))
		if err != nil {
			fmt.Println("Erreur lors de l'envoi de la demande POST à l'API pour /getUserPoints:", err)
			return
		}
		defer respGetUserPoints.Body.Close()

		if respGetUserPoints.Status == "200 OK" {
			fmt.Printf("Points de l'utilisateur : %s\n", userName)
		} else {
			fmt.Printf("Échec de la récupération des points de l'utilisateur : %s\n", userName)
		}
	}
}

func main() {
	ip := "10.49.122.144"

	startPort := 1024
	endPort := 65535

	var wg sync.WaitGroup

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		go checkPort(ip, port, &wg)
	}

	wg.Wait()
}

func isConnectionError(err error) bool {

	if ne, ok := err.(net.Error); ok && ne != nil {
		if ne.Temporary() || ne.Timeout() {
			return true
		}
	}
	return false
}
