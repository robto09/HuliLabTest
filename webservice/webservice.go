package webservice

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

func ConvertDollarsColones() string{
	rtrn := ""
	var currentTime = time.Now()
	var indicator = "318"
	var startDate = fmt.Sprint(currentTime.Day(), "/", currentTime.Month(), "/", currentTime.Year())
	var endDate = fmt.Sprint(currentTime.Day(), "/", currentTime.Month(), "/", currentTime.Year())
	var nombre = "Maria"
	var subNiveles = "N"
	var correoElectronico = "mariaobando09@gmail.com"
	var token = "0OOIOO49MA"
	var baseUrl = "https://gee.bccr.fi.cr/Indicadores/Suscripciones/WS/wsindicadoreseconomicos.asmx/ObtenerIndicadoresEconomicosXML"
	var url = baseUrl +"?Indicador="+indicator+"&FechaInicio="+ startDate +"&FechaFinal="+ endDate +"&nombre="+ nombre +"&subNiveles="+ subNiveles +"&correoElectronico="+ correoElectronico +"&Token="+token

	resp, err := http.Get(url)

	if err != nil {
		fmt.Println(err)
		return rtrn
	}
	// fmt.Println(resp)
	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Println(err2)
		return rtrn
	}
	xmlData := string(body)
	// fmt.Println(data)
	re := regexp.MustCompile(`&lt;NUM_VALOR&gt;(.*)&lt;/NUM_VALOR&gt;`)
	NumvalorTag := re.FindString(xmlData)
	NumvalorRightTag := strings.Replace(NumvalorTag, "&lt;NUM_VALOR&gt;", "", -1)
	rate := strings.Replace(NumvalorRightTag, "&lt;/NUM_VALOR&gt;", "", -1)
	return rate
}
