package main

import (
	"bufio"
	"cmp"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// код писать тут

const testToken = "test-token"

var (
	requestValid = &SearchRequest{
		Limit:      6,
		Offset:     4,
		Query:      "Amet",
		OrderBy:    -1,
		OrderField: "Name",
	}
	requestValid100 = &SearchRequest{
		Limit:      100,
		Offset:     0,
		Query:      "",
		OrderBy:    1,
		OrderField: "Id",
	}
)

type TestCase struct {
	NameCase    string
	AccessToken string
	Request     *SearchRequest
	Result      *SearchResponse
	IsError     bool
}

type Row struct {
	Text          string `xml:"-"`
	ID            int    `xml:"id,int"`
	Guid          string `xml:"-"`
	IsActive      string `xml:"-"`
	Balance       string `xml:"-"`
	Picture       string `xml:"-"`
	Age           int    `xml:"age,int"`
	EyeColor      string `xml:"-"`
	FirstName     string `xml:"first_name,string"`
	LastName      string `xml:"last_name,string"`
	Gender        string `xml:"gender,string"`
	Company       string `xml:"-"`
	Email         string `xml:"-"`
	Phone         string `xml:"-"`
	Address       string `xml:"-"`
	About         string `xml:"about,string"`
	Registered    string `xml:"-"`
	FavoriteFruit string `xml:"-"`
}

type Root struct {
	XMLName xml.Name `xml:"-"`
	Text    string   `xml:"-"`
	Row     []Row    `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != "test-token" {
		w.Header().Set("WWW-Authenticate", `Basic realm="api"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := r.URL.Query()
	query := strings.ToLower(params.Get("query"))

	orderField := params.Get("order_field")
	switch orderField {
	case "Id", "Age", "Name", "":
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SearchErrorResponse{Error: ErrorBadOrderField})
		return
	}

	orderBy, err := strconv.Atoi(params.Get("order_by"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch orderBy {
	case OrderByAsc, OrderByAsIs, OrderByDesc:
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(SearchErrorResponse{Error: "OrderBy invalid"})
		return
	}

	limit, err := strconv.Atoi(params.Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if limit > 25 {
		limit = 25
	}

	offset, err := strconv.Atoi(params.Get("offset"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if offset != 0 {
		offset--
	}

	file, err := os.Open("dataset.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	dec := xml.NewDecoder(bufio.NewReader(file))
	var name strings.Builder

	users := make([]User, 0, 1024)

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "row" {
			continue
		}

		var row Row
		if err := dec.DecodeElement(&row, &se); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		name.Reset()
		_, err = name.WriteString(row.FirstName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		_, err = name.WriteString(" ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		_, err = name.WriteString(row.LastName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if strings.Contains(strings.ToLower(row.About), query) ||
			strings.Contains(strings.ToLower(name.String()), query) {
			users = append(users,
				User{
					Id:     row.ID,
					Name:   name.String(),
					Age:    row.Age,
					About:  row.About,
					Gender: row.Gender,
				})

		}
	}

	slices.SortFunc(users, func(a, b User) int {
		var order int
		switch orderField {
		case "Id":
			order = cmp.Compare(a.Id, b.Id)
		case "Age":
			order = cmp.Compare(a.Age, b.Age)
		case "Name":
			order = strings.Compare(a.Name, b.Name)
		default:
			order = strings.Compare(a.Name, b.Name)
		}
		return order * orderBy
	})

	if offset > len(users) {
		offset = len(users)
	}

	end := offset + limit
	if limit > len(users) {
		end = len(users)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users[offset:end]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SearchServerErrorJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{broken json`))
}

func SearchServerBrokenResultJSON(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{broken json`))
}

func SearchServerFatalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("empty"))
}

func SearchServerTimeOut(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	SearchServer(w, r)
}

func SearchServerClientUnknownError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			NameCase:    "Limit less than zero",
			AccessToken: testToken,
			Request:     &SearchRequest{Limit: -1},
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			NameCase:    "Offset less than zero",
			AccessToken: testToken,
			Request:     &SearchRequest{Offset: -1},
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			NameCase:    "Unauthorized",
			AccessToken: "invalid-token",
			Request:     &SearchRequest{},
			Result:      nil,
			IsError:     true,
		},
		TestCase{
			NameCase:    "OrderBy invalid",
			AccessToken: testToken,
			Request: &SearchRequest{
				Limit:      6,
				Offset:     4,
				Query:      "Amet",
				OrderBy:    2,
				OrderField: "Name",
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			NameCase:    "OrderField invalid",
			AccessToken: testToken,
			Request: &SearchRequest{
				Limit:      6,
				Offset:     4,
				Query:      "Amet",
				OrderBy:    -1,
				OrderField: "Gender",
			},
			Result:  nil,
			IsError: true,
		},
		TestCase{
			NameCase:    "ValidCase",
			AccessToken: testToken,
			Request:     requestValid,
			Result: &SearchResponse{
				Users: []User{
					{Id: 27, Name: "Rebekah Sutton", Age: 26, About: "Aliqua exercitation ad nostrud et exercitation amet quis cupidatat esse nostrud proident. Ullamco voluptate ex minim consectetur ea cupidatat in mollit reprehenderit voluptate labore sint laboris. Minim cillum et incididunt pariatur amet do esse. Amet irure elit deserunt quis culpa ut deserunt minim proident cupidatat nisi consequat ipsum.\n", Gender: "female"},
					{Id: 31, Name: "Palmer Scott", Age: 37, About: "Elit fugiat commodo laborum quis eu consequat. In velit magna sit fugiat non proident ipsum tempor eu. Consectetur exercitation labore eiusmod occaecat adipisicing irure consequat fugiat ullamco aliquip nostrud anim irure enim. Duis do amet cillum eiusmod eu sunt. Minim minim sunt sit sit enim velit sint tempor enim sint aliquip voluptate reprehenderit officia. Voluptate magna sit consequat adipisicing ut eu qui.\n", Gender: "male"},
					{Id: 4, Name: "Owen Lynn", Age: 30, About: "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n", Gender: "male"},
					{Id: 14, Name: "Nicholson Newman", Age: 23, About: "Tempor minim reprehenderit dolore et ad. Irure id fugiat incididunt do amet veniam ex consequat. Quis ad ipsum excepteur eiusmod mollit nulla amet velit quis duis ut irure.\n", Gender: "male"},
					{Id: 20, Name: "Lowery York", Age: 27, About: "Dolor enim sit id dolore enim sint nostrud deserunt. Occaecat minim enim veniam proident mollit Lorem irure ex. Adipisicing pariatur adipisicing aliqua amet proident velit. Magna commodo culpa sit id.\n", Gender: "male"},
					{Id: 7, Name: "Leann Travis", Age: 34, About: "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n", Gender: "female"},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			NameCase:    "ValidCaseLimit25",
			AccessToken: testToken,
			Request:     requestValid100,
			Result: &SearchResponse{
				Users: []User{
					{Id: 0, Name: "Boyd Wolf", Age: 22, About: "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n", Gender: "male"},
					{Id: 1, Name: "Hilda Mayer", Age: 21, About: "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n", Gender: "female"},
					{Id: 2, Name: "Brooks Aguilar", Age: 25, About: "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n", Gender: "male"},
					{Id: 3, Name: "Everett Dillard", Age: 27, About: "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n", Gender: "male"},
					{Id: 4, Name: "Owen Lynn", Age: 30, About: "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n", Gender: "male"},
					{Id: 5, Name: "Beulah Stark", Age: 30, About: "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n", Gender: "female"},
					{Id: 6, Name: "Jennings Mays", Age: 39, About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n", Gender: "male"},
					{Id: 7, Name: "Leann Travis", Age: 34, About: "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n", Gender: "female"},
					{Id: 8, Name: "Glenn Jordan", Age: 29, About: "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n", Gender: "male"},
					{Id: 9, Name: "Rose Carney", Age: 36, About: "Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n", Gender: "female"},
					{Id: 10, Name: "Henderson Maxwell", Age: 30, About: "Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n", Gender: "male"},
					{Id: 11, Name: "Gilmore Guerra", Age: 32, About: "Labore consectetur do sit et mollit non incididunt. Amet aute voluptate enim et sit Lorem elit. Fugiat proident ullamco ullamco sint pariatur deserunt eu nulla consectetur culpa eiusmod. Veniam irure et deserunt consectetur incididunt ad ipsum sint. Consectetur voluptate adipisicing aute fugiat aliquip culpa qui nisi ut ex esse ex. Sint et anim aliqua pariatur.\n", Gender: "male"},
					{Id: 12, Name: "Cruz Guerrero", Age: 36, About: "Sunt enim ad fugiat minim id esse proident laborum magna magna. Velit anim aliqua nulla laborum consequat veniam reprehenderit enim fugiat ipsum mollit nisi. Nisi do reprehenderit aute sint sit culpa id Lorem proident id tempor. Irure ut ipsum sit non quis aliqua in voluptate magna. Ipsum non aliquip quis incididunt incididunt aute sint. Minim dolor in mollit aute duis consectetur.\n", Gender: "male"},
					{Id: 13, Name: "Whitley Davidson", Age: 40, About: "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n", Gender: "male"},
					{Id: 14, Name: "Nicholson Newman", Age: 23, About: "Tempor minim reprehenderit dolore et ad. Irure id fugiat incididunt do amet veniam ex consequat. Quis ad ipsum excepteur eiusmod mollit nulla amet velit quis duis ut irure.\n", Gender: "male"},
					{Id: 15, Name: "Allison Valdez", Age: 21, About: "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n", Gender: "male"},
					{Id: 16, Name: "Annie Osborn", Age: 35, About: "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n", Gender: "female"},
					{Id: 17, Name: "Dillard Mccoy", Age: 36, About: "Laborum voluptate sit ipsum tempor dolore. Adipisicing reprehenderit minim aliqua est. Consectetur enim deserunt incididunt elit non consectetur nisi esse ut dolore officia do ipsum.\n", Gender: "male"},
					{Id: 18, Name: "Terrell Hall", Age: 27, About: "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n", Gender: "male"},
					{Id: 19, Name: "Bell Bauer", Age: 26, About: "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n", Gender: "male"},
					{Id: 20, Name: "Lowery York", Age: 27, About: "Dolor enim sit id dolore enim sint nostrud deserunt. Occaecat minim enim veniam proident mollit Lorem irure ex. Adipisicing pariatur adipisicing aliqua amet proident velit. Magna commodo culpa sit id.\n", Gender: "male"},
					{Id: 21, Name: "Johns Whitney", Age: 26, About: "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n", Gender: "male"},
					{Id: 22, Name: "Beth Wynn", Age: 31, About: "Proident non nisi dolore id non. Aliquip ex anim cupidatat dolore amet veniam tempor non adipisicing. Aliqua adipisicing eu esse quis reprehenderit est irure cillum duis dolor ex. Laborum do aute commodo amet. Fugiat aute in excepteur ut aliqua sint fugiat do nostrud voluptate duis do deserunt. Elit esse ipsum duis ipsum.\n", Gender: "female"},
					{Id: 23, Name: "Gates Spencer", Age: 21, About: "Dolore magna magna commodo irure. Proident culpa nisi veniam excepteur sunt qui et laborum tempor. Qui proident Lorem commodo dolore ipsum.\n", Gender: "male"},
					{Id: 24, Name: "Gonzalez Anderson", Age: 33, About: "Quis consequat incididunt in ex deserunt minim aliqua ea duis. Culpa nisi excepteur sint est fugiat cupidatat nulla magna do id dolore laboris. Aute cillum eiusmod do amet dolore labore commodo do pariatur sit id. Do irure eiusmod reprehenderit non in duis sunt ex. Labore commodo labore pariatur ex minim qui sit elit.\n", Gender: "male"},
				},
				NextPage: false,
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: item.AccessToken,
			URL:         ts.URL,
		}
		result, err := c.FindUsers(*item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(item.Result, result) {
			t.Errorf("[%d, %s] wrong result, expected %#v, got %#v", caseNum, item.NameCase, item.Result, result)
		}
	}
}

func TestFindUsersErrorJSON(t *testing.T) {
	item := TestCase{
		NameCase:    "ErrorJSON",
		AccessToken: testToken,
		Request:     requestValid,
		Result:      nil,
		IsError:     true,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServerErrorJSON))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: testToken,
		URL:         ts.URL,
	}
	result, err := c.FindUsers(*item.Request)
	if err != nil && !item.IsError {
		t.Errorf("unexpected error: %#v", err)
	}
	if err == nil && item.IsError {
		t.Errorf("expected error, got nil")
	}
	if !reflect.DeepEqual(item.Result, result) {
		t.Errorf("[%s] wrong result, expected %#v, got %#v", item.NameCase, item.Result, result)
	}
}

func TestFindUsersBrokenResultJSON(t *testing.T) {
	item := TestCase{
		NameCase:    "BrokenResultJSON",
		AccessToken: testToken,
		Request:     requestValid,
		Result:      nil,
		IsError:     true,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServerBrokenResultJSON))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: testToken,
		URL:         ts.URL,
	}
	result, err := c.FindUsers(*item.Request)
	if err != nil && !item.IsError {
		t.Errorf("unexpected error: %#v", err)
	}
	if err == nil && item.IsError {
		t.Errorf("expected error, got nil")
	}
	if !reflect.DeepEqual(item.Result, result) {
		t.Errorf("[%s] wrong result, expected %#v, got %#v", item.NameCase, item.Result, result)
	}
}

func TestFindUsersFatalError(t *testing.T) {
	item := TestCase{
		NameCase:    "ErrorJSON",
		AccessToken: testToken,
		Request:     requestValid,
		Result:      nil,
		IsError:     true,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServerFatalError))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: testToken,
		URL:         ts.URL,
	}
	result, err := c.FindUsers(*item.Request)
	if err != nil && !item.IsError {
		t.Errorf("unexpected error: %#v", err)
	}
	if err == nil && item.IsError {
		t.Errorf("expected error, got nil")
	}
	if !reflect.DeepEqual(item.Result, result) {
		t.Errorf("[%s] wrong result, expected %#v, got %#v", item.NameCase, item.Result, result)
	}
}

func TestFindUsersTimeOut(t *testing.T) {
	item := TestCase{
		NameCase:    "TimeOut",
		AccessToken: testToken,
		Request:     requestValid,
		Result:      nil,
		IsError:     true,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServerTimeOut))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: testToken,
		URL:         ts.URL,
	}
	result, err := c.FindUsers(*item.Request)
	if err != nil && !item.IsError {
		t.Errorf("unexpected error: %#v", err)
	}
	if err == nil && item.IsError {
		t.Errorf("expected error, got nil")
	}
	if !reflect.DeepEqual(item.Result, result) {
		t.Errorf("[%s] wrong result, expected %#v, got %#v", item.NameCase, item.Result, result)
	}
}

func TestFindUsersClientUnknownError(t *testing.T) {
	item := TestCase{
		NameCase:    "ClientUnknownError",
		AccessToken: testToken,
		Request:     requestValid,
		Result:      nil,
		IsError:     true,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServerClientUnknownError))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: testToken,
		URL:         ts.URL,
	}
	result, err := c.FindUsers(*item.Request)
	if err != nil && !item.IsError {
		t.Errorf("unexpected error: %#v", err)
	}
	if err == nil && item.IsError {
		t.Errorf("expected error, got nil")
	}
	if !reflect.DeepEqual(item.Result, result) {
		t.Errorf("[%s] wrong result, expected %#v, got %#v", item.NameCase, item.Result, result)
	}
}
