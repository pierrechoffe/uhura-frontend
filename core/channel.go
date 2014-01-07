package core

type Channel struct {
  Title         string
  Description   string
  ImageUrl      string
  Copyright     string
  LastBuildDate string
  Url           string `sql:"not null;unique"`
  Id            int
}
