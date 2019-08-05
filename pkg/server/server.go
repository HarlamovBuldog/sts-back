package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/HarlamovBuldog/social-tournament-service/pkg/sts"
	"github.com/gorilla/mux"
)

// Service is the wrapper for all methods working with db.
type Service interface {
	// AddUser adds user to db with given name, auto-increments id
	// and sets user balance to zero. Returns userID if succeed.
	AddUser(ctx context.Context, name string) (string, error)

	// GetUser returns *User that contains all information about user with help of provided id
	GetUser(ctx context.Context, id string) (*storage.User, error)

	DeleteUser(ctx context.Context, id string) error

	// TakeUserBalance finds user with provided id and deducts from his balance provided points
	TakeUserBalance(ctx context.Context, id string, points float64) error

	// FundUserBalance finds user with provided id and adds to his balance provided points
	FundUserBalance(ctx context.Context, id string, points float64) error

	AddTournament(ctx context.Context, name string, deposit float64) (string, error)
	GetTournament(ctx context.Context, id string) (*storage.Tournament, error)
	DeleteTournament(ctx context.Context, id string) error
	CalculateTournamentPrize(ctx context.Context, id string) error
	SetTournamentWinner(ctx context.Context, tournamentID, userID string) error
	AddUserToTournamentList(ctx context.Context, tournamentID, userID string) error
}
type Server struct {
	http.Handler
	service sts.Service
}

// NewServer constructs a Server, decodes yaml configuration file
// and assigns decoded values to config struct.
func NewServer(db sts.Service) *Server {
	router := mux.NewRouter()

	s := Server{
		service: db,
		Handler: router,
	}
	router.HandleFunc("/user", s.createNewUser).Methods("POST")
	router.HandleFunc("/user/{id}", s.getUserInfo).Methods("GET")
	router.HandleFunc("/user/{id}", s.removeUser).Methods("DELETE")
	router.HandleFunc("/user/{id}/take", s.takeUserBonusPoints).Methods("POST")
	router.HandleFunc("/user/{id}/fund", s.addUserBonusPoints).Methods("POST")
	return &s
}

func (s *Server) createNewUser(w http.ResponseWriter, req *http.Request) {
	var user sts.User
	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error creating new user: can't decode request body: %v\n", err)
		return
	}
	userID, err := s.service.AddUser(req.Context(), user.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error creating new user: %v\n", err)
		return
	}
	err = json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{
		ID: userID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error creating new user: error encoding json: %v\n", err)
		return
	}
}

func (s *Server) getUserInfo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userID, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error getting user info: user id is not provided")
		return
	}

	userData, err := s.service.GetUser(req.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error getting user info: %v\n", err)
		return
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	err = enc.Encode(&userData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error getting user info: error encoding json: %v\n", err)
		return
	}
}

func (s *Server) removeUser(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userID, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error removing user: user id is not provided")
		return
	}
	err := s.service.DeleteUser(req.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error removing user info: %v\n", err)
		return
	}
}

func (s *Server) takeUserBonusPoints(w http.ResponseWriter, req *http.Request) {
	points := struct {
		Points float64 `json:"points"`
	}{}
	err := json.NewDecoder(req.Body).Decode(&points)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error taking user bonus points: can't decode request body: %v\n", err)
		return
	}

	vars := mux.Vars(req)
	userID, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error taking user bonus points: user id is not provided")
		return
	}

	err = s.service.TakeUserBalance(req.Context(), userID, points.Points)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error taking user bonus points: %v", err)
		return
	}
}

func (s *Server) addUserBonusPoints(w http.ResponseWriter, req *http.Request) {
	points := struct {
		Points float64 `json:"points"`
	}{}
	err := json.NewDecoder(req.Body).Decode(&points)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error adding user bonus points: can't decode request body: %v\n", err)
		return
	}

	vars := mux.Vars(req)
	userID, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error adding user bonus points: user id is not provided")
		return
	}

	err = s.service.FundUserBalance(req.Context(), userID, points.Points)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error adding user bonus points: %v", err)
		return
	}
}