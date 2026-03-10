import * as vscode from 'vscode';
import { execFile } from 'child_process';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
} from 'vscode-languageclient/node';

let client: LanguageClient;

const scheme = 'shdoc-preview';

class ShdocPreviewProvider implements vscode.TextDocumentContentProvider {
    private _onDidChange = new vscode.EventEmitter<vscode.Uri>();
    readonly onDidChange = this._onDidChange.event;

    refresh(uri: vscode.Uri) {
        this._onDidChange.fire(uri);
    }

    provideTextDocumentContent(uri: vscode.Uri): Thenable<string> {
        const sourcePath = uri.query;
        return new Promise((resolve) => {
            execFile('shdoc-ng', ['generate', '--format', 'markdown', '-i', sourcePath], (err, stdout, _stderr) => {
                if (err) {
                    resolve(`# Error\n\n\`\`\`\n${err.message}\n\`\`\``);
                } else {
                    resolve(stdout);
                }
            });
        });
    }
}

export function activate(context: vscode.ExtensionContext) {
    const serverOptions: ServerOptions = {
        command: 'shdoc-ng',
        args: ['lsp'],
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [
            { scheme: 'file', language: 'shellscript' },
            { scheme: 'file', language: 'sh' },
            { scheme: 'file', language: 'bash' },
        ],
    };

    client = new LanguageClient(
        'shdoc-ng',
        'shdoc-ng',
        serverOptions,
        clientOptions,
    );
    client.start();
    context.subscriptions.push(client);

    // Documentation preview
    const provider = new ShdocPreviewProvider();
    context.subscriptions.push(
        vscode.workspace.registerTextDocumentContentProvider(scheme, provider),
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('shdoc-ng.previewDocumentation', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                return;
            }

            const sourceUri = editor.document.uri;
            const previewUri = vscode.Uri.parse(
                `${scheme}:${sourceUri.path}.md?${sourceUri.fsPath}`,
            );

            await vscode.commands.executeCommand(
                'markdown.showPreviewToSide',
                previewUri,
            );

            // Refresh preview when the source document changes.
            const watcher = vscode.workspace.onDidChangeTextDocument((e) => {
                if (e.document.uri.toString() === sourceUri.toString()) {
                    provider.refresh(previewUri);
                }
            });
            context.subscriptions.push(watcher);
        }),
    );
}

export function deactivate(): Thenable<void> | undefined {
    return client?.stop();
}
