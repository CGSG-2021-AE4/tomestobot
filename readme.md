## Deployment 
- setup env
- `go build -o bot ./cmd/`
- `bot`

## Envionment variables

- `BX_DOMAIN` - domain for bitrix api(like hostname.bitrix.ru)
- `BX_USER_ID` - id of bitrix user which will be used for API requests
- `BX_HOOK` - bitrix hook for API requests
- `TG_TOKEN` - telegram bot token
- `ENABLE_DEBUG_LOGS` - enable debug level logs flag(enable if `true`)
- `ENABLE_RESTY_LOGS` - enable resty level logs flag(enable if `true`)
- `ID_STORE_FILE` - name json file for known users id storage
- `ADMIN_WHITELIST` - list of usernames of telegram users which will receive logs(are splited only by spaces)

## Some description
### Tagged var
Since tg allow only 64 bytes of payload I have to store dynamic data localy.
In msg payload I leave tag(uuid) to ensure the data is valid and the one I want.
Different handlers expect different data:
- onStart - nothing
- onListDeals - nothing
- onDealActions - deals tag, id
- onWriteComment - deal tag
- onAddComment - deal tag(exception because I can use it's dynamic data at all)
- onListTasks - deal tag,
- onCompleteTask - deal tasks tag, id
Based on what handler want it expects the first 16(or 32 in onCompleteTask case) of payload to be uuid.
