package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/savioxavier/termlink"
)

var (
	helpStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("122"))
	batmanStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#18ff00"))
	bannerStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#5fd760"))
	unselectedListStyle     = lipgloss.NewStyle().Margin(1, 1).Border(lipgloss.RoundedBorder())
	unselectedViewportStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).Margin(1, 2, 0, 0)
	selectedListStyle       = lipgloss.NewStyle().Margin(1, 1).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#37d900"))
	selectedViewportStyle   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1).Margin(1, 2, 0, 0).BorderForeground(lipgloss.Color("#37d900"))
)

const (
	host = "localhost"
	port = "22"
)

type sessionState uint
type page uint
type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	state    sessionState
	page     page
	list     list.Model
	viewport viewport.Model
	ready    bool
	cache    [5]string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) changePage() {
	m.page = page(m.list.Index())
	switch m.page {
	case 0:
		m.viewport.SetContent(m.getHome())
	case 1:
		m.viewport.SetContent(m.getAbout())
	case 2:
		m.viewport.SetContent(m.cache[2])
	case 3:
		m.viewport.SetContent(m.getContact())
	case 4:
		m.viewport.SetContent(m.cache[4])
	}
}

func (m *model) changeFocus() {
	if m.state == 0 {
		m.state = 1
	} else if m.state == 1 {
		m.state = 0
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" {
			return m, tea.Quit
		}
		if k := msg.String(); k == "enter" || k == "right" {
			m.changePage()
			m.changeFocus()
		}
		if k := msg.String(); k == "left" {
			if m.state == 1 {
				m.changeFocus()
			}
		}
	case tea.WindowSizeMsg:
		h, v := unselectedListStyle.GetFrameSize()
		m.list.SetSize(((24 * msg.Width) / 100), msg.Height-v)
		m.list.SetShowStatusBar(false)
		m.list.SetShowHelp(false)

		if !m.ready {
			m.viewport = viewport.New((65*msg.Width)/100, msg.Height-v)
			m.viewport.YPosition = h

			m.ready = true
			m.viewport.Style = lipgloss.NewStyle()

			banner := `
			~~~~~~~~~~~~~~~~~~~~~~~~
			404 ameyaNotFound.lol
		  ~~~~~~~~~~~~~~~~~~~~~~~~`
			banner = lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, bannerStyle.Render(banner))
			banner += "\n \n \n"
			batman := `
			..oo$00ooo..                    ..ooo00$oo..
			.o$$$$$$$$$,                          ,$$$$$$$$$o.
		 .o$$$$$$$$$,             .   .              ,$$$$$$$$$o.
	   .o$$$$$$$$$$~             /$   $\              ~$$$$$$$$$$o.
	 .{$$$$$$$$$$$.              $\___/$               .$$$$$$$$$$$}.
	o$$$$$$$$$$$$8              .$$$$$$$.               8$$$$$$$$$$$$o
   $$$$$$$$$$$$$$$              $$$$$$$$$               $$$$$$$$$$$$$$$
  o$$$$$$$$$$$$$$$.             o$$$$$$$o              .$$$$$$$$$$$$$$$o
  $$$$$$$$$$$$$$$$$.           o{$$$$$$$}o            .$$$$$$$$$$$$$$$$$
 ^$$$$$$$$$$$$$$$$$$.         J$$$$$$$$$$$L          .$$$$$$$$$$$$$$$$$$^
 !$$$$$$$$$$$$$$$$$$$$oo..oo$$$$$$$$$$$$$$$$$oo..oo$$$$$$$$$$$$$$$$$$$$$!
 {$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$}
 6$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$?
 ,$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$,
  o$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$o
   $$$$$$$$$$$$$$;,~,^Y$$$7^,,o$$$$$$$$$$$o,,^Y$$$7^,~,;$$$$$$$$$$$$$$$
   ,$$$$$$$$$$$,       ,$,    ,,$$$$$$$$$,     ,$,       ,$$$$$$$$$$$$,
	!$$$$$$$$$7         !       ,$$$$$$$,       !         V$$$$$$$$$!
	 ^o$$$$$$!                   ,$$$$$,                   !$$$$$$o^
	   ^$$$$$,                    $$$$$                    ,$$$$$^
		 ,o$$$,                   ^$$$,                   ,$$$o,
		   ~$$$.                   $$$.                  .$$$~
			 ,$;.                  ,$,                  .;$,
				,.                  !                  ., `
			batman = lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, batmanStyle.Render(batman))
			batman += "\n \n \n"

			c := lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, helpStyle.Render("Navigation: Arrow Keys + Enter \nQuit: Ctrl + C"))
			text := banner + batman + c
			text = lipgloss.PlaceVertical(20, lipgloss.Center, text)
			m.viewport.SetContent(text)

		} else {
			m.viewport.Width = (75 * msg.Width) / 100
			m.viewport.Height = msg.Height - v
		}

	}

	var cmd tea.Cmd
	if m.state == 0 {
		m.list, cmd = m.list.Update(msg)

	} else if m.state == 1 {
		m.viewport, cmd = m.viewport.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	var s string
	s1 := unselectedListStyle.Render(m.list.View())
	s2 := unselectedViewportStyle.Render(m.viewport.View())

	if m.state == 0 {
		s1 = selectedListStyle.Render(m.list.View())
		s2 = unselectedViewportStyle.Render(m.viewport.View())
	} else if m.state == 1 {
		s1 = unselectedListStyle.Render(m.list.View())
		s2 = selectedViewportStyle.Render(m.viewport.View())
	}

	s = lipgloss.JoinHorizontal(0, s1, s2)
	return s
}

func newModel() model {
	items := []list.Item{
		item{title: "Home"},
		item{title: "About Me"},
		item{title: "Projects"},
		item{title: "Contact"},
		item{title: "Certifications"},
	}
	d := list.NewDefaultDelegate()
	c := lipgloss.Color("#008741")
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(c).BorderLeftForeground(c)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(c).BorderLeftForeground(c)
	m := model{state: 0, page: 0, list: list.New(items, d, 0, 0)}

	m.list.Title = "Ameya Taneja"
	// Set TitleBar style without highlights
	m.list.Styles.Title = lipgloss.NewStyle().
		Padding(1, 1, 1, 1).
		Foreground(lipgloss.Color("#37d900")).
		Background(lipgloss.Color("0"))
	m.cache[2] = m.getProjects()
	m.cache[4] = m.getCertifications()
	return m
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	m := newModel()
	return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
}

var (
	headingStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00")).MarginTop(1)
	boxStyle            = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#008741")).Background(lipgloss.Color("#1E1E1E")).Width(48) // Dark background with a hacker color
	boxDescriptionStyle = lipgloss.NewStyle().Width(48).MarginTop(1).Foreground(lipgloss.Color("#A0FFA0"))
	boxTechStyle        = lipgloss.NewStyle().Width(48).Foreground(lipgloss.Color("#00FF00")).MarginTop(1).Bold(true)
	boxHeadingStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#008741")).Margin(0)
	mainColour          = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	width               = lipgloss.NewStyle().Width(44).Foreground(lipgloss.Color("#cdcdcd"))
	bold                = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	email               = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffa500"))
	github              = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF00FF"))
	whatsapp            = lipgloss.NewStyle().Foreground(lipgloss.Color("#5FF475")).Bold(true)
	linkedin            = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFFF")).Bold(true)
)

type box struct {
	title        string
	description  string
	technologies string
	link         string
}

type projects []box

func (b *box) getStr() string {
	text := boxHeadingStyle.Render(b.title)
	desc := boxDescriptionStyle.Render(b.description)
	text = lipgloss.JoinVertical(0, text, desc)
	tech := fmt.Sprintf("%s %s", boxTechStyle.Render("Techonologies: "), b.technologies)
	text = lipgloss.JoinVertical(0, text, tech)
	link := termlink.ColorLink("Github Repository", b.link, "blue")
	text = lipgloss.JoinVertical(0, text, link)
	text = boxStyle.Render(text)
	return text
}

type certiBox struct {
	title            string
	issued_by        string
	certification_id string
	issued_date      string
}
type certifications []certiBox

func (c *certiBox) getStr() string {
	title := boxHeadingStyle.Render(c.title)
	issuedBy := boxDescriptionStyle.Render(fmt.Sprintf("Issued By: %s", c.issued_by))
	text := lipgloss.JoinVertical(0, title, issuedBy)
	certID := boxTechStyle.Render(fmt.Sprintf("Certification ID: %s", c.certification_id))
	issuedDate := boxTechStyle.Render(fmt.Sprintf("Issued Date: %s", c.issued_date))
	text = lipgloss.JoinVertical(0, text, certID, issuedDate)
	text = boxStyle.Render(text)
	return text

}
func (m *model) getHome() string {
	banner := `
~~~~~~~~~~~~~~~~~~~~~~~~
  404 ameyaNotFound.lol
~~~~~~~~~~~~~~~~~~~~~~~~`
	banner = lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, bannerStyle.Render(banner))
	banner += "\n \n \n"
	batman := `
                   ..oo$00ooo..                    ..ooo00$oo..
                .o$$$$$$$$$,                          ,$$$$$$$$$o.
             .o$$$$$$$$$,             .   .              ,$$$$$$$$$o.
           .o$$$$$$$$$$~             /$   $\              ~$$$$$$$$$$o.
         .{$$$$$$$$$$$.              $\___/$               .$$$$$$$$$$$}.
        o$$$$$$$$$$$$8              .$$$$$$$.               8$$$$$$$$$$$$o
       $$$$$$$$$$$$$$$              $$$$$$$$$               $$$$$$$$$$$$$$$
      o$$$$$$$$$$$$$$$.             o$$$$$$$o              .$$$$$$$$$$$$$$$o
      $$$$$$$$$$$$$$$$$.           o{$$$$$$$}o            .$$$$$$$$$$$$$$$$$
     ^$$$$$$$$$$$$$$$$$$.         J$$$$$$$$$$$L          .$$$$$$$$$$$$$$$$$$^
     !$$$$$$$$$$$$$$$$$$$$oo..oo$$$$$$$$$$$$$$$$$oo..oo$$$$$$$$$$$$$$$$$$$$$!
     {$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$}
     6$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$?
     ,$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$,
      o$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$o
       $$$$$$$$$$$$$$;,~,^Y$$$7^,,o$$$$$$$$$$$o,,^Y$$$7^,~,;$$$$$$$$$$$$$$$
       ,$$$$$$$$$$$,       ,$,    ,,$$$$$$$$$,     ,$,       ,$$$$$$$$$$$$,
        !$$$$$$$$$7         !       ,$$$$$$$,       !         V$$$$$$$$$!
         ^o$$$$$$!                   ,$$$$$,                   !$$$$$$o^
           ^$$$$$,                    $$$$$                    ,$$$$$^
             ,o$$$,                   ^$$$,                   ,$$$o,
               ~$$$.                   $$$.                  .$$$~
                 ,$;.                  ,$,                  .;$,
                    ,.                  !                  .,     `
	batman = lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, batmanStyle.Render(batman))
	batman += "\n \n \n"
	c := lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, helpStyle.Render("Navigation: Arrow Keys + Enter \nQuit: Ctrl + C or q"))
	text := banner + batman + c
	text = lipgloss.PlaceVertical(20, lipgloss.Center, text)
	return text
}

func (m *model) getAbout() string {

	t1 := bold.Render(mainColour.Render("Ameya Taneja\n"))
	t2 := "Results-driven B.Tech CSE student specializing in cybersecurity and forensics, equipped with a solid technical foundation in full-stack development. Eager to apply programming skills in JavaScript, Python, and SQL to drive impactful solutions and contribute to innovative projects within a dynamic team environment."
	t2 = width.Render(t2)
	t1 = lipgloss.JoinVertical(0, t1, t2)
	t1 = lipgloss.JoinVertical(lipgloss.Center, t1)
	t1 = lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, t1)
	t1 = lipgloss.PlaceVertical(m.viewport.Height, lipgloss.Center, t1)
	return t1
}

func (m *model) getContact() string {
	t1 := bold.Render(mainColour.Render("Contact me:"))
	t2 := helpStyle.Render("You can reach me here\n")
	t1 = lipgloss.JoinVertical(0, t1, t2)
	t2 = github.Render("Github: ") + termlink.Link("@realtneu", "https://github.com/realTNEU")
	t1 = lipgloss.JoinVertical(0, t1, t2)
	t2 = email.Render("Email: ") + termlink.Link("ameyataneja2015@gmail.com", "mailto:ameyataneja@gmail.com")
	t1 = lipgloss.JoinVertical(0, t1, t2)
	t2 = whatsapp.Render("Whatsapp: ") + termlink.Link("+91 8368015613", "https://api.whatsapp.com/send/?phone=918368015613&text&type=phone_number&app_absent=0")
	t1 = lipgloss.JoinVertical(0, t1, t2)
	t2 = linkedin.Render("Linkedin: ") + termlink.Link("@ameyataneja", "https://www.linkedin.com/in/ameyataneja")
	t1 = lipgloss.JoinVertical(0, t1, t2)
	t1 = lipgloss.PlaceHorizontal(m.viewport.Width, lipgloss.Center, t1)
	t1 = lipgloss.PlaceVertical(m.viewport.Height, lipgloss.Center, t1)
	return t1
}

func (m *model) getProjects() string {
	text := "Here are some of the projects I have worked on.\n"
	text = headingStyle.Render(text)

	p := projects{
		box{
			title:        "connectify",
			description:  "Connectify is a dynamic web application designed to facilitate seamless video calling and chatting experiences. This platform aims to enhance communication by providing users with an intuitive interface for real-time interaction. Users can engage in one-on-one or group video calls and exchange text messages.",
			technologies: "MongoDb, React.js, Express.js and Node.Js",
			link:         "https://github.com/realTNEU/connectify"},
		box{
			title:        "Ip Vulnerability Tracker",
			description:  "The Vulnerability Tracker project architecture is designed to deliver comprehensive insights into website security and performance by integrating multiple scanning tools and APIs. The system adopts a modular approach to ensure scalability, reliability, and ease of integration with various external services.",
			technologies: "React.js, Express.js, CSS and Node.js",
			link:         "https://github.com/Akshat-NegI27/IP-Track"},
		box{
			title:        "CORS-API-PROXY",
			description:  "An API proxy server to block CORS block",
			technologies: "Express.js",
			link:         "https://github.com/realTNEU/api-proxy-express"},
		box{
			title:        "tneuGPT",
			description:  "tneuGPT is an uncensored GPT run on a locally run model i.e. Dolphin Llama v3 with a functional frontend.",
			technologies: "Express.js, LM studio, Node.js",
			link:         "https://github.com/realTNEU/tneuGPT"},
	}

	for _, b := range p {
		text = lipgloss.JoinVertical(0, text, b.getStr())
	}
	return text
}
func (m *model) getCertifications() string {
	text := "Certifications\n"
	text = headingStyle.Render(text)
	c := certifications{
		certiBox{
			title:            "Advanced React ",
			issued_by:        "Meta",
			issued_date:      "July 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/3LYLE7WQVCJ2"},
		certiBox{
			title:            "Programming in Python ",
			issued_by:        "Meta",
			issued_date:      "July 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/BB6BTSJDDAEP"},
		certiBox{
			title:            "APIs ",
			issued_by:        "Meta",
			issued_date:      "July 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/BE775XH9W438"},
		certiBox{
			title:            "Programming with JavaScript ",
			issued_by:        "Meta",
			issued_date:      "June 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/JVHXDENAQ3QB"},
		certiBox{
			title:            "HTML and CSS in depth ",
			issued_by:        "Meta",
			issued_date:      "June 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/86FYQ9AR8P7L"},
		certiBox{
			title:            "Version Control ",
			issued_by:        "Meta",
			issued_date:      "June 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/U9GB5EGDSHCA"},
		certiBox{
			title:            "Principles of UX/UI design",
			issued_by:        "Meta",
			issued_date:      "July 2024",
			certification_id: "https://www.coursera.org/account/accomplishments/verify/47S8NRY82K5W"},
	}
	for _, b := range c {
		text = lipgloss.JoinVertical(0, text, b.getStr())
	}
	return text

}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_rsa"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
