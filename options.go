package ipfscliwrapper

import (
	"time"

	"github.com/bartmika/ipfs-cli-wrapper/internal/oskit"
	"github.com/bartmika/ipfs-cli-wrapper/internal/randomkit"
	"github.com/bartmika/ipfs-cli-wrapper/internal/urlkit"
)

// WithContinousOperation is a functional option to configure our wrapper to
// not terminate the operation of the `ipfs` binary when running in the
// background; in addition when we run the `Start()` function, no errors will
// occure pertaining to previously active running `ipfs` binary instance. This
// is a useful option if you developing an app in which you restart often and
// you don't want to restart the `ipfs` binary often then use this option.
func WithContinousOperation() Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.isDaemonRunningContinously = true
	}
}

// WithOverrideBinaryOsAndArch is a functional option to configure our wrapper
// to use a specific binary. The available `os` options are: darwin, linux,
// freebsd, openbsd and windows. The available `arch` choices are: arm, arm64,
// 386, and amd64.
func WithOverrideBinaryOsAndArch(overrideOS, overrideArch string) Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.os = overrideOS
		wrap.arch = overrideArch
	}
}

// WithOverrideDaemonInitialWarmupDuration is a functional option to configure
// our wrapper to set a custom warmup delay for our app to give a custom delay
// to allow the `ipfs` to loadup before giving your app execution control.
func WithOverrideDaemonInitialWarmupDuration(seconds int) Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.daemonInitialWarmupDuration = time.Duration(seconds) * time.Second
	}
}

// WithForcedShutdownDaemonOnStartup is a functional option to add if you want
// this package to look for any previously running `ipfs` binary in the system
// background and shut it down before our package loads up a new `ipfs` binary
// instance.
func WithForcedShutdownDaemonOnStartup() Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.forceShutdownOnStartup = true
	}
}

// WithDenylist is a functional option which downloads a `denylist` [0] from the
// URL you provided and applies it to the `ipfs` binary running instance.
// [0] https://github.com/ipfs/kubo/blob/master/docs/content-blocking.md
func WithDenylist(denylistFilename string, denylistURL string) Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.denylistFilename = denylistFilename
		wrap.denylistURL = denylistURL
	}
}

func WithCustomOsOperator(osOperator oskit.OSOperater) Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.osOperator = osOperator
	}
}

func WithCustomUrlDownloader(urlDownloader urlkit.URLDownloader) Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.urlDownloader = urlDownloader
	}
}

func WithCustomRandomGenerator(gen randomkit.RandomGenerator) Option {
	return func(wrap *ipfsCliWrapper) {
		wrap.randomGenerator = gen
	}
}
