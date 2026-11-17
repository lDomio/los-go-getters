package httpUtils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

//*=====================================| SERVER |========================================================\\

func StartServer(puerto int, mux *http.ServeMux) {

	log.Println("Se prendio el servidor")
	//Aca se convierte el puerto a string
	puertoString := strconv.Itoa(puerto)
	//Agrego el : que es necesario para empezar los servers
	//Los puertos tienen que ser strings, enviamos int y se convierte a string
	err := http.ListenAndServe(":"+puertoString, mux)
	if err != nil {
		panic(err)
	}
	log.Println("Se apago el servidor")

}

//*=====================================| CLIENT |========================================================\\

// retorna true si la request fue exitosa, false en caso contrario
func RequestBody(port int, ip string, metodo string, query string, body []byte) (*http.Response, error) {
	// Crear un cliente HTTP
	cliente := &http.Client{}

	// Construir la URL
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, query)

	// Crear la solicitud HTTP
	req, err := http.NewRequest(metodo, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creando request a ip: %s puerto: %d: %w", ip, port, err)
	}

	// Establecer los headers
	req.Header.Set("Content-Type", "application/json")

	// Enviar la solicitud al servidor
	respuesta, err := cliente.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error enviando request a ip: %s puerto: %d: %w", ip, port, err)
	}

	// Verificar el código de estado de la respuesta
	if respuesta.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status error: %d", respuesta.StatusCode)
	}

	// Todo salió bien
	return respuesta, nil
}

func EnviarPOST(ip string, port int, query string, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, query)
	respuesta, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error enviando POST a %s: %w", url, err)
	}
	return respuesta, err
}

func EnviarGET(ip string, port int, query string) (*http.Response, error) {
	url := fmt.Sprintf("http://%s:%d/%s", ip, port, query)
	respuesta, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error enviando GET a %s: %w", url, err)
	}
	return respuesta, err
}

func CodificarJSON(data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error codificando a JSON: %w", err)
	}
	return body, nil
}

func DecodificarJSON(w http.ResponseWriter, r *http.Request, target interface{}) error {
	// Decodifica en formato JSON la request.
	err := json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		log.Println("Error al decodificar el JSON:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return fmt.Errorf("error decodificando JSON: %w", err)
	}
	return nil
}
