package evmosswagger

import (
	_ "github.com/evmos/evmos/v20/client/docs/statik" // Import Evmos statik
	"github.com/rakyll/statik/fs"
)

// https://github.com/rakyll/statik/issues/56

// FS is the Evmos swagger filesystem
var FS, _ = fs.New()
