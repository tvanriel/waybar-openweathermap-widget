/*
Copyright © 2024 Ted van Riel

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/briandowns/openweathermap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var cfgFile string
var (
	icons = map[string]string{
		"01d": "☀️",
		"02d": "⛅️",
		"03d": "☁️",
		"04d": "☁️",
		"09d": "🌧️",
		"10d": "🌦️",
		"11d": "⛈️",
		"13d": "🌨️",
		"50d": "🌫",

		"01n": "☀️",
		"02n": "⛅️",
		"03n": "☁️",
		"04n": "☁️",
		"09n": "🌧️",
		"10n": "🌦️",
		"11n": "⛈️",
		"13n": "🌨️",
		"50n": "🌫",
	}

	timefmt = "15:04 MST"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "waybar-openweathermap [lat] [long] [key]",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(3),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		long, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			log.Fatalf("parse longitude %s: %v", args[0], err)
		}
		
                lat, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			log.Fatalf("parse latitude %s: %v", args[1], err)
		}

		request, err := openweathermap.NewCurrent(
                        "C", 
                        "en", 
                        args[2], 
                )
		
                if err != nil {
			log.Fatalf("get weatherinfo: %v", err)
                        return
		}

		err = request.CurrentByCoordinates(
			&openweathermap.Coordinates{
				Longitude: long,
				Latitude:  lat,
			},
		)

                if err != nil {
                        log.Fatalf("get weather: %v", err)
                        return
                }

		temp := request.Main.Temp
		icon := icons[request.Weather[0].Icon]
		desc := description(request.Weather[0].ID)

		feelsLike := request.Main.FeelsLike
		humidity := request.Main.Humidity
		pressure := request.Main.Pressure
		sunrise := time.Unix(int64(request.Sys.Sunrise), 0).Format(timefmt)
		sunset := time.Unix(int64(request.Sys.Sunset), 0).Format(timefmt)
		windSpeed := request.Wind.Speed

		data := &result{
			Text: text(icon, strconv.FormatFloat(temp, 'f', 1, 64)),
			Tooltip: tooltip(
                                desc,
				strconv.FormatInt(int64(feelsLike), 10),
				strconv.FormatInt(int64(pressure), 10),
				strconv.FormatInt(int64(humidity), 10),
				sunrise,
				sunset,
				strconv.FormatFloat(windSpeed, 'f', 0, 64),
			),
                        Class: "weather",
		}

		b, err := json.Marshal(data)
		if err != nil {
			log.Fatalf("encode json: %v", err)
		}
		os.Stdout.Write(b)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.waybar-openweathermap.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".waybar-openweathermap" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".waybar-openweathermap")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func text(icon, temp string) string {
	return strings.Join([]string{
		icon,
                " ",
		temp,
		" °C",
	}, "")
}
func tooltip(desc, feelsLike, pressure, humidity, sunrise, sunset, windspeed string) string {
        caser := cases.Title(language.German)

	return strings.TrimLeftFunc(strings.Join([]string{
                caser.String(desc),
                "\n",
		"Feels like ",
		feelsLike,
		" °C\n",
		"Pressure ",
		pressure,
		" hPa\n",
		"Humidity ",
		humidity,
		"%\n",
		"Sunrise ",
		sunrise,
		"\n",
		"Sunset ",
		sunset,
		"\n",
		"Wind speed ",
		windspeed,
		" m/sec",
	}, ""), unicode.IsSpace)
}

type result struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
}


func description(id int) (string) {
        groups := [][]*openweathermap.ConditionData{
                openweathermap.ThunderstormConditions,
                openweathermap.DrizzleConditions,
                openweathermap.RainConditions,
                openweathermap.SnowConditions,
                openweathermap.AtmosphereConditions,
                openweathermap.CloudConditions,
                openweathermap.AdditionalConditions,
        }
        for g := range groups {
                for c := range groups[g] {
                        if groups[g][c].ID == id {
                                return groups[g][c].Meaning
                        }
                }
        }
        return ""
}
