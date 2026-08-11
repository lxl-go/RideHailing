import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'

const source = fs.readFileSync(path.resolve('src/pages/orderChat/orderChat.vue'), 'utf8')

assert.match(source, /<u-button[\s\S]*>发送<\/u-button>/, 'send button should render with a valid closing tag')
assert.match(source, /class="message-row"[\s\S]*:class="\{ mine: isMine\(item\) \}"/, 'message rows should use normalized mine detection')
assert.match(source, /class="bubble mine-bubble"|\.mine \.bubble/, 'own messages should have right-side bubble styling')
assert.match(source, /class="composer"/, 'chat page should render bottom composer')
assert.match(source, /placeholder="输入消息"/, 'composer should use Chinese placeholder')
assert.match(source, /connectionText/, 'chat page should show websocket connection state')
assert.match(source, /client_msg_id/, 'sent messages should include client_msg_id for dedupe')
assert.doesNotMatch(source, /:disabled="!connected \|\| !draft\.trim\(\)"/, 'send button should not be blocked forever while websocket is reconnecting')
assert.match(source, /pendingMessages/, 'chat page should queue messages while websocket is not open')
assert.match(source, /flushPendingMessages/, 'queued messages should be flushed after websocket opens')
assert.match(source, /connectionStatus/, 'chat page should distinguish connecting, connected, and failed states')

console.log('passenger order chat contract tests passed')
