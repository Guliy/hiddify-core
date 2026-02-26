package mobile

import (
	"encoding/json"
	"os"

	hcore "github.com/hiddify/hiddify-core/v2/hcore"
	"github.com/hiddify/hiddify-core/v2/config"

	_ "net/http/pprof"

	_ "github.com/sagernet/gomobile"
	"github.com/sagernet/sing-box/experimental/libbox"
)

type SetupOptions struct {
	BasePath        string
	WorkingDir      string
	TempDir         string
	Listen          string
	Secret          string
	Debug           bool
	Mode            int
	FixAndroidStack bool
}

func Setup(opt *SetupOptions, platformInterface libbox.PlatformInterface) error {
	return hcore.Setup(&hcore.SetupRequest{
		BasePath:          opt.BasePath,
		WorkingDir:        opt.WorkingDir,
		TempDir:           opt.TempDir,
		FlutterStatusPort: 0,
		Listen:            opt.Listen,
		Debug:             opt.Debug,
		Mode:              hcore.SetupMode(opt.Mode),
		Secret:            opt.Secret,
		FixAndroidStack:   opt.FixAndroidStack,
	}, platformInterface)

	// return hcore.Start(17078)
}

// func Start(configPath string, configContent string, platformInterface libbox.PlatformInterface) (*hcore.CoreInfoResponse, error) {
// 	state, err := hcore.StartWithPlatformInterface(&hcore.StartRequest{
// 		ConfigContent: configContent,
// 		ConfigPath:    configPath,
// 	}, platformInterface)
// 	return state, err
// }

func Start(configPath string, configContent string) error {
	_, err := hcore.StartService(libbox.BaseContext(nil), &hcore.StartRequest{
		ConfigPath:    configPath,
		ConfigContent: configContent,
	})
	return err
}

func Stop() error {
	_, err := hcore.Stop()
	return err
}

func GetServerPublicKey() []byte {
	return hcore.GetGrpcServerPublicKey()
}

func AddGrpcClientPublicKey(clientPublicKey []byte) error {
	return hcore.AddGrpcClientPublicKey(clientPublicKey)
}

func Close(mode int) {
	hcore.Close(hcore.SetupMode(mode))
}

func Test() string {
	return "Hello from mobile"
}

func Pause() {
	hcore.Pause()
}

func Wake() {
	hcore.Wake()
}

// ParseConfig reads a proxy config from tempPath (supports vless://, vmess://, sing-box JSON, clash YAML),
// converts it to a full sing-box JSON configuration with TUN enabled (required for iOS VPN),
// and writes the result to configPath. optionsJson is a JSON string of HiddifyOptions (may be empty).
// Returns an empty string on success, or an error message on failure.
func ParseConfig(tempPath string, configPath string, optionsJson string) string {
	ctx := libbox.BaseContext(nil)

	opts := config.DefaultHiddifyOptions()
	opts.InboundOptions.EnableTun = true // iOS VPN extension requires TUN inbound

	if optionsJson != "" {
		var parsed config.HiddifyOptions
		if err := json.Unmarshal([]byte(optionsJson), &parsed); err == nil {
			parsed.InboundOptions.EnableTun = true
			opts = &parsed
		}
	}

	configBytes, err := config.ParseBuildConfigBytes(ctx, opts, &config.ReadOptions{Path: tempPath})
	if err != nil {
		return err.Error()
	}
	if err := os.WriteFile(configPath, configBytes, 0o644); err != nil {
		return err.Error()
	}
	return ""
}
