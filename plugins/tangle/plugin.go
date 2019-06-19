package tangle

import (
	"github.com/iotaledger/goshimmer/packages/node"
)

// region plugin module setup //////////////////////////////////////////////////////////////////////////////////////////

var PLUGIN = node.NewPlugin("Tangle", configure, run)

func configure(plugin *node.Plugin) {
	configureTransactionDatabase(plugin)
	configureTransactionMetaDataDatabase(plugin)
	configureApproversDatabase(plugin)
	configureSolidifier(plugin)
}

func run(plugin *node.Plugin) {
	// this plugin has no background workers
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
