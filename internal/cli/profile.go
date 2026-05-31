package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Profile struct {
	Name          string `mapstructure:"name"`
	BaseURL       string `mapstructure:"base_url"`
	DefaultOutput string `mapstructure:"default_output"`
	Active        bool   `mapstructure:"active"`
}

type Session struct {
	ProfileName  string `mapstructure:"profile_name"`
	Username     string `mapstructure:"username"`
	Token        string `mapstructure:"token"`
	ExpireAt     string `mapstructure:"expire_at"`
	LastVerified string `mapstructure:"last_verified_at"`
}

func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".octopus")
	}
	return filepath.Join(home, ".octopus")
}

func configPath() string {
	return filepath.Join(configDir(), "config.yaml")
}

func initViper() {
	viper.SetConfigFile(configPath())
	viper.SetConfigType("yaml")
	_ = os.MkdirAll(configDir(), 0755)
	_ = viper.ReadInConfig()
}

func saveViper() {
	_ = os.MkdirAll(configDir(), 0755)
	_ = viper.WriteConfigAs(configPath())
}

func GetProfiles() []Profile {
	initViper()
	var profiles []Profile
	_ = viper.UnmarshalKey("profiles", &profiles)
	return profiles
}

func AddProfile(name, baseURL, defaultOutput string) error {
	initViper()
	profiles := GetProfiles()
	for _, p := range profiles {
		if p.Name == name {
			return nil
		}
	}
	profiles = append(profiles, Profile{
		Name:          name,
		BaseURL:       baseURL,
		DefaultOutput: defaultOutput,
		Active:        len(profiles) == 0,
	})
	viper.Set("profiles", profiles)
	saveViper()
	return nil
}

func SetActiveProfile(name string) error {
	initViper()
	profiles := GetProfiles()
	for i := range profiles {
		profiles[i].Active = profiles[i].Name == name
	}
	viper.Set("profiles", profiles)
	saveViper()
	return nil
}

func RemoveProfile(name string) error {
	initViper()
	profiles := GetProfiles()
	filtered := make([]Profile, 0, len(profiles))
	for _, p := range profiles {
		if p.Name != name {
			filtered = append(filtered, p)
		}
	}
	viper.Set("profiles", filtered)
	saveViper()
	return nil
}

func GetActiveProfile() *Profile {
	profiles := GetProfiles()
	for _, p := range profiles {
		if p.Active {
			return &p
		}
	}
	if len(profiles) > 0 {
		p := profiles[0]
		p.Active = true
		SetActiveProfile(p.Name)
		return &p
	}
	return nil
}

func GetSession() *Session {
	initViper()
	var session Session
	err := viper.UnmarshalKey("session", &session)
	if err != nil || session.ProfileName == "" {
		return nil
	}
	return &session
}

func SaveSession(session Session) {
	initViper()
	viper.Set("session", session)
	saveViper()
}

func ClearSession() {
	initViper()
	viper.Set("session", nil)
	saveViper()
}
