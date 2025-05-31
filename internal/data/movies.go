package data

import (
    "time"
)


type Movie struct {
    ID        int64     `json:"id"`
    CreatedAt time.Time `json:"-"` // Use the - directive
    Title     string    `json:"title"`
    Year      int32     `json:"year,omitzero"`    // Add the omitzero directive
    Runtime   Runtime     `json:"runtime,omitzero"` // Add the omitzero directive
    Genres    []string  `json:"genres,omitzero"`  // Add the omitzero directive
    Version   int32     `json:"version"`
}
