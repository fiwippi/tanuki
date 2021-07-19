package core

import (
	"fmt"
	"strconv"
	"time"
)

type SeriesMetadata struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	DateReleased *Date  `json:"date_released"`
}

func (m *SeriesMetadata) String() string {
	return fmt.Sprintf("Title: %s, Author: %s, DateReleased: %s",
		m.Title, m.Author, m.DateReleased)
}

func NewSeriesMetadata() *SeriesMetadata {
	return &SeriesMetadata{
		DateReleased: NewDate(time.Time{}),
	}
}

type EntryMetadata struct {
	Title        string `json:"title"`
	Chapter      int    `json:"chapter"`
	Volume       int    `json:"volume"`
	Author       string `json:"author"`
	DateReleased *Date  `json:"date_released"`
}

func (m *EntryMetadata) String() string {
	return fmt.Sprintf("Title: %s, Chapter: %s, Volume: %s, DateReleased: %s",
		m.Title, m.StringChapter(), m.StringVolume(), m.DateReleased)
}

func (m *EntryMetadata) StringChapter() string {
	if m.Chapter == -1 {
		return "N/A"
	}
	return strconv.Itoa(m.Chapter)
}

func (m *EntryMetadata) StringVolume() string {
	if m.Volume == -1 {
		return "N/A"
	}
	return strconv.Itoa(m.Volume)
}

func NewEntryMetadata() *EntryMetadata {
	return &EntryMetadata{
		DateReleased: NewDate(time.Time{}),
	}
}