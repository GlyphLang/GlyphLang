"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.deactivate = exports.activate = void 0;
const vscode_1 = require("vscode");
const node_1 = require("vscode-languageclient/node");
let client;
function activate(context) {
    // Get configuration
    const config = vscode_1.workspace.getConfiguration('glyph');
    const lspPath = config.get('lsp.path') || 'glyph';
    const logFile = config.get('lsp.logFile') || '';
    // Build server command
    const serverArgs = ['lsp'];
    if (logFile) {
        serverArgs.push('--log', logFile);
    }
    const serverOptions = {
        command: lspPath,
        args: serverArgs,
        transport: node_1.TransportKind.stdio
    };
    // Options to control the language client
    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'glyph' }],
        synchronize: {
            fileEvents: vscode_1.workspace.createFileSystemWatcher('**/*.{glyph,glybc}')
        }
    };
    // Create the language client
    client = new node_1.LanguageClient('glyphLanguageServer', 'Glyph Language Server', serverOptions, clientOptions);
    // Start the client (this will also launch the server)
    client.start();
}
exports.activate = activate;
function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
exports.deactivate = deactivate;
//# sourceMappingURL=extension.js.map