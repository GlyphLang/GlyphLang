import * as path from 'path';
import { workspace, ExtensionContext } from 'vscode';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
    // Get configuration
    const config = workspace.getConfiguration('glyph');
    const lspPath = config.get<string>('lsp.path') || 'glyph';
    const logFile = config.get<string>('lsp.logFile') || '';

    // Build server command
    const serverArgs = ['lsp'];
    if (logFile) {
        serverArgs.push('--log', logFile);
    }

    const serverOptions: ServerOptions = {
        command: lspPath,
        args: serverArgs,
        transport: TransportKind.stdio
    };

    // Options to control the language client
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'glyph' }],
        synchronize: {
            fileEvents: workspace.createFileSystemWatcher('**/*.{glyph,glybc}')
        }
    };

    // Create the language client
    client = new LanguageClient(
        'glyphLanguageServer',
        'Glyph Language Server',
        serverOptions,
        clientOptions
    );

    // Start the client (this will also launch the server)
    client.start();
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    return client.stop();
}
