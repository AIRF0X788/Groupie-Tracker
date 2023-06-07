package main

import (
	"C"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// On importe les packages nécessaires.
import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// On définit les structures qui seront utilisées pour stocker les données des artistes, des concerts passés et futurs,
// ainsi que les relations entre les artistes.

var client *http.Client

type Relations struct {
	Relations []DatesLocation `json:"index"`
}
type DatesLocation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}
type Concert struct {
	Location string
	Dates    string
}

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"concertDates"`
	RelationsUrl string   `json:"relations"`
	PastConcert  []Concert
	FuturConcert []Concert
}

// Fonction pour télécharger l'image d'un artiste.
func downloadImage(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", "image")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		panic(err)
	}

	return file.Name()
}

// Fonction pour récupérer la liste de tous les artistes.
func getAllArtists() []Artist {
	var artists []Artist
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		return nil
	}

	return artists
}

// Fonction pour récupérer les relations entre les artistes et leurs concerts passés/futurs.
func getRelations(artists []Artist) []Artist {
	var relations Relations
	var newArtists []Artist = artists

	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/relation")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&relations)
	if err != nil {
		panic(err)
	}

	for _, element := range relations.Relations {
		for city, dates := range element.DatesLocations {
			for _, dateString := range dates {
				date, err := time.Parse("02-01-2006", dateString)
				if err != nil {
					panic(err)
				}
				if date.Before(time.Now()) {
					newArtists[element.ID-1].PastConcert = append(newArtists[element.ID-1].PastConcert, Concert{city, dateString})
				} else {
					newArtists[element.ID-1].FuturConcert = append(newArtists[element.ID-1].PastConcert, Concert{city, dateString})
				}
			}
		}
	}

	return newArtists
}

func main() {
	// Récupérer la liste de tous les artistes
	var artists []Artist = getAllArtists()

	// Initialiser des listes pour stocker les noms des artistes,
	// les concerts passés et les concerts à venir
	var listArtistName []string
	var listConcertPast []string
	var listConcertFutur []string

	// Récupérer les relations entre les artistes
	getRelations(artists)

	// Créer une nouvelle application
	a := app.New()

	// Créer une nouvelle fenêtre avec un titre
	w := a.NewWindow("Projet Groupie")

	// Définir la taille de la fenêtre
	w.Resize(fyne.NewSize(1000, 700))

	// Créer un menu principal avec un bouton quitter
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Quitter"),

		// Créer un menu principal avec deux sous-menus : "Thème sombre" et "Thème clair"
		fyne.NewMenu("Thème",
			fyne.NewMenuItem("Thème sombre", func() {
				a.Settings().SetTheme(theme.DarkTheme())
			}),
			fyne.NewMenuItem("Thème clair", func() {
				a.Settings().SetTheme(theme.LightTheme())
			}),
		),

		// Créer un menu principal avec deux sous-menus : "Spotify" et "Concert"
		fyne.NewMenu("Envie de plus ?",
			fyne.NewMenuItem("Spotify", func() {
				// Ouvrir une URL Spotify dans le navigateur par défaut
				u, _ := url.Parse("https://open.spotify.com/")
				_ = a.OpenURL(u)
			}),
			fyne.NewMenuItem("Concert", func() {
				// Ouvrir une URL Ticketmaster dans le navigateur par défaut
				u, _ := url.Parse("https://www.ticketmaster.fr/fr/concert/")
				_ = a.OpenURL(u)
			}),
		))

	// Ajouter le menu principal à la fenêtre
	w.SetMainMenu(mainMenu)

	// Créer des widgets pour afficher les informations sur l'artiste sélectionné
	artist := artists[0]
	aTitle := widget.NewLabelWithStyle("Info Artist :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	aTitle.Move(fyne.NewPos(520, 0))

	aName := widget.NewLabel(" ")

	aMembers := widget.NewLabelWithStyle(" ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Italic: true})
	aMembers.Move(fyne.NewPos(520, 250))

	aImage := canvas.NewImageFromFile("")
	aImage.Resize(fyne.NewSize(150, 190))
	aImage.Move(fyne.NewPos(520, 50))

	aCreationDate := widget.NewLabel(" ")
	aFirstAlbum := widget.NewLabel(" ")
	aLabelPastConcert := widget.NewLabelWithStyle("Concerts passés :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	aLabelPastConcert.Move(fyne.NewPos(520, 450))
	aLabelFuturConcert := widget.NewLabelWithStyle("Concerts à venir :", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	aLabelFuturConcert.Move(fyne.NewPos(750, 450))

	// Créer une nouvelle liste pour les concerts passés
	aPastConcerts := widget.NewList(
		func() int { return 1 },                                 // nombre initial d'éléments
		func() fyne.CanvasObject { return widget.NewLabel("") }, // élément par défaut
		func(lii widget.ListItemID, co fyne.CanvasObject) { // fonction pour mettre à jour les éléments
		},
	)
	aPastConcerts.Resize(fyne.NewSize(250, 200)) // ajuster la taille de la liste
	aPastConcerts.Move(fyne.NewPos(520, 500))    // positionner la liste sur la fenêtre

	// Créer une nouvelle liste pour les concerts à venir
	aFuturConcerts := widget.NewList(
		func() int { return 1 },                                 // nombre initial d'éléments
		func() fyne.CanvasObject { return widget.NewLabel("") }, // élément par défaut
		func(lii widget.ListItemID, co fyne.CanvasObject) { // fonction pour mettre à jour les éléments
		},
	)
	aFuturConcerts.Resize(fyne.NewSize(250, 200)) // ajuster la taille de la liste
	aFuturConcerts.Move(fyne.NewPos(750, 500))    // positionner la liste sur la fenêtre

	// Ajouter les noms des artistes à la liste
	for _, artist := range artists {
		listArtistName = append(listArtistName, artist.Name)
	}

	// Créer une liste des noms des artistes
	list := widget.NewList(
		func() int { return len(listArtistName) },                                 // nombre d'éléments
		func() fyne.CanvasObject { return widget.NewLabel("Liste des artistes") }, // élément par défaut
		func(lii widget.ListItemID, co fyne.CanvasObject) { // fonction pour mettre à jour les éléments
			co.(*widget.Label).SetText(listArtistName[lii]) // mettre à jour le texte de l'élément avec le nom de l'artiste correspondant
		},
	)
	list.Resize(fyne.NewSize(280, 500)) // ajuster la taille de la liste
	list.Move(fyne.NewPos(220, 0))      // positionner la liste sur la fenêtre

	// Créer une zone de recherche pour les artistes
	searchEntry := widget.NewEntry()
	searchButton := widget.NewButton("Rechercher", func() {
		// Nouvelle liste pour les résultats de la recherche
		filteredList := []string{}

		// Parcourir la liste d'origine et ajouter les éléments correspondants à la nouvelle liste
		for _, item := range listArtistName {
			if strings.Contains(strings.ToLower(item), strings.ToLower(searchEntry.Text)) {
				filteredList = append(filteredList, item)
			}
		}

		// Mettre à jour la liste avec les résultats de la recherche
		list.Length = func() int {
			return len(filteredList)
		}
		list.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}
		list.UpdateItem = func(index int, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(filteredList[index])
		}
		list.Refresh()
	})
	clearButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		searchEntry.SetText("")

		// Mettre à jour la liste avec les résultats de la recherche
		list.Length = func() int {
			return len(listArtistName)
		}
		list.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}
		list.UpdateItem = func(index int, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(listArtistName[index])
		}
		list.Refresh()
	})

	// Création d'une barre de recherche contenant une entrée de recherche, un bouton de recherche et un bouton de réinitialisation
	searchBar := container.NewVBox(
		searchEntry,
		searchButton,
		clearButton,
	)
	searchBar.Resize(fyne.NewSize(200, 100))

	// Création d'un conteneur pour le contenu principal de l'application
	content := container.NewWithoutLayout(
		searchBar,
		list,
	)

	// Création d'un séparateur et positionnement à l'écran
	separator := widget.NewSeparator()
	separator.Move(fyne.NewPos(500, 0))

	// Création d'un conteneur pour afficher les informations de l'artiste sélectionné
	nameContainer := container.NewVBox(
		aName,
		aFirstAlbum,
	)
	nameContainer.Move(fyne.NewPos(750, 50))

	// Création d'un conteneur pour les informations de l'artiste
	infoArtist := container.NewWithoutLayout(
		aTitle,
		aImage,
		nameContainer,
		aMembers,
		aPastConcerts,
		aFuturConcerts,
		aLabelPastConcert,
		aLabelFuturConcert,
	)
	infoArtist.Resize(fyne.NewSize(480, 800))
	// infoArtist.Move(fyne.NewPos(520, 0))
	infoArtist.Hide()

	// Lorsqu'un élément de la liste est sélectionné
	list.OnSelected = func(id widget.ListItemID) {

		// Récupération de l'artiste correspondant à l'élément sélectionné
		artist = artists[id]

		// Mise à jour des éléments d'interface utilisateur avec les informations de l'artiste
		aName.Text = artist.Name
		aName.Refresh()

		aMembersList := "Membre : \n"
		for _, member := range artist.Members {
			aMembersList += "- " + member + "\n"
		}
		aMembers.Text = aMembersList
		aMembers.Refresh()

		aCreationDate.Text = strconv.Itoa(artist.CreationDate)

		aImagePath := downloadImage(artist.Image)
		aImage.File = aImagePath
		aImage.Refresh()

		aFirstAlbum.Text = "Date première album : " + artist.FirstAlbum
		aFirstAlbum.Refresh()

		// Mise à jour de la liste des concerts passés de l'artiste sélectionné
		listConcertPast = nil
		for _, concert := range artists[id].PastConcert {
			listConcertPast = append(listConcertPast, concert.Location+" : "+concert.Dates)
		}
		aPastConcerts.Length = func() int {
			return len(listConcertPast)
		}
		aPastConcerts.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}
		aPastConcerts.UpdateItem = func(index int, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(listConcertPast[index])
		}
		aPastConcerts.Refresh()

		// Mise à jour de la liste des concerts futurs de l'artiste sélectionné
		listConcertFutur = nil
		for _, concert := range artists[id].PastConcert {
			listConcertFutur = append(listConcertFutur, concert.Location+" : "+concert.Dates)
		}
		aPastConcerts.Length = func() int {
			return len(listConcertFutur)
		}
		aPastConcerts.CreateItem = func() fyne.CanvasObject {
			return widget.NewLabel("")
		}

		aPastConcerts.Refresh()

		if infoArtist.Hidden {
			infoArtist.Show()
		}

	}

	w.SetContent(
		container.NewWithoutLayout(
			content,
			separator,
			infoArtist),
	)

	w.ShowAndRun()
}
