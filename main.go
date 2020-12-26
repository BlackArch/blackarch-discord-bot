package main

import (
    "github.com/bwmarrin/discordgo"
    "flag"
    "strings"
    "os/signal"
    "os"
    "syscall"
    "fmt"
    "net/http"
    "time"
    "io/ioutil"
    "bufio"
)

type buffer_t struct {
    content string
    timestamp int64
}

type entry_t struct {
    name string
    version string
    description string
    group string
    url string
}

var (
    tool_buffer buffer_t
    entry_list []entry_t
)

func prettyOutput(text string) (*discordgo.MessageEmbed) {
    thumbnail := discordgo.MessageEmbedThumbnail{
        URL: "https://blackarch.org/images/logo/ba-logo.png",
        Width: 50,
        Height: 50,
    }
    
    output := discordgo.MessageEmbed{
        Title: "BlackArch Tool Search",
        URL: "https://blackarch.org/tools.html",
        Thumbnail: &thumbnail,
        Color: 13369344,
        Description: text,
    }

    return &output
}

func bot_log(m *discordgo.MessageCreate, trigger string) {
    fmt.Printf("GuildID: %s, ChannelID: %sm User: %s, Time: %s, Trigger: %s.\n",
        m.Message.GuildID, m.Message.ChannelID, m.Message.Author,
        m.Message.Timestamp, trigger)
}

func createToolBuffer() (err error) {
    res, err := http.Get("https://raw.githubusercontent.com/BlackArch/blackarch-site/master/data/tools")
    if err != nil {
        return
    }

    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return
    }

    entry_list = nil

    tool_buffer.content = fmt.Sprintf("%s", body)
    tool_buffer.timestamp = (time.Now()).Unix()

    scanner := bufio.NewScanner(strings.NewReader(tool_buffer.content))
    for scanner.Scan() {
        var entry_tmp entry_t
        var tokenized []string = strings.Split(scanner.Text(), "|")

        entry_tmp.name = tokenized[0]
        entry_tmp.version = tokenized[1]
        entry_tmp.description = tokenized[2]
        entry_tmp.group = tokenized[3]
        entry_tmp.url = tokenized[4]

        entry_list = append(entry_list, entry_tmp)
    }

    return
}

func toolListUpdate() (result string, err error) {
    now := (time.Now()).Unix()
    if (now - 1800) < tool_buffer.timestamp {
        result = "Too soon, can't update yet."
        return
    }

    err = createToolBuffer()
    result = "Updated!"

    return
}

func searchTool(search string) (output string, err error) {
    template := "Name: %s\nVersion: %s\nDescription: %s\nGroup: %s\nURL: %s\n\n"

    for _, element := range entry_list {
        if strings.Contains(element.name, search) {
            output += fmt.Sprintf(template, element.name, element.version,
                element.description, element.group, element.url)

            if len(output) >= 1000 {
                output += "\n-> **TRUNCATED**: too big, change your search."
                break
            }
        }
    }

    if len(output) <= 0 {
        output = "Sorry, nothing found"
    }

    return
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore all messages created by the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }

    // Tokenize message for parsing
    var tokenized []string = strings.Split(m.Content, " ")

    switch tokenized[0] {
    case "+ping":
        s.ChannelMessageSendEmbed(m.ChannelID, prettyOutput("pong"))

    case "+search":
        // Needs a parameters to search for
        if len(tokenized[0:]) < 2 {
            s.ChannelMessageSendEmbed(m.ChannelID,
                prettyOutput("Are you crazy? Search what?"))
            return
        }

        output, err := searchTool(tokenized[1])
        if err != nil {
            s.ChannelMessageSendEmbed(m.ChannelID,
                prettyOutput("Something is wrong. Get a hold of an admin!"))
            return
        }

        s.ChannelMessageSendEmbed(m.ChannelID, prettyOutput(output))

    case "+update":
        output, _ := toolListUpdate()
        s.ChannelMessageSendEmbed(m.ChannelID, prettyOutput(output))
    }
}

func main() {
    // Get bot token. Maybe not a good idea?!
    bot_token := flag.String("token", "", "The bot token.")
    flag.Parse()

    if len(*bot_token) <= 0 {
        fmt.Println("You must be nutts! Where is my token?")
        return
    }

    err := createToolBuffer()
    if err != nil {
        fmt.Println("Arghhh, can't get the tool list!")
        return
    }

    // Create a new Discord session using the provided bot token.
    dg, err := discordgo.New("Bot " + *bot_token)
    if err != nil {
        fmt.Println("Something is fucked up when connecting to Discord: ", err)
        return
    }

    // Register the events callback handlers.
    dg.AddHandler(messageHandler)

    // Open a websocket connection to Discord and begin listening.
    err = dg.Open()
    if err != nil {
        fmt.Println("Something is fucked up when openning connection: ", err)
        return
    }

    fmt.Println("We are on, baby... :)")

    // Wait here until CTRL-C or other term signal is received.
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc

    // Cleanly close down the Discord session.
    fmt.Println("Bye-bye, darling... :(")
    dg.Close()
}

