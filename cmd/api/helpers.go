package main

import (
    "errors"
    "net/http"
    "strconv"
    "encoding/json"
    "github.com/julienschmidt/httprouter"
)

type Envelope map[string]any


func (app *application) readIDParam(r *http.Request) (int64, error) {
  params := httprouter.ParamsFromContext(r.Context())

  id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
  if err != nil || id < 1 {
    return 0, errors.New("invalid id parameter")
  }

  return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data Envelope, headers http.Header) error {
  
  js, err := json.MarshalIndent(data, "", "\t")
  if err != nil {
    return err
  }
  js = append(js, '\n')
  for key, value := range headers {
    w.Header()[key] = value
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(js)

  return nil
}
