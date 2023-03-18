# [Nostr Citadel - The Sovereign Relay | WiP](https://github.com/MrJohnsson77/nostr-citadel)  

Nostr Citadel, a personal relay that stores and safeguards all your nostr data.  

The idea is that anyone should be able to run a relay and manage their own data in a **simple** and convenient way.
Once you're up and running, the relay will actively sync data from other relays and create a backup locally.

By default, only admin data is synced to the relay.  
Unless set as public, only admin and whitelisted npubs are allowed to post to the relay.

Stay Sovereign 🤙

![Nostr Citadel](screenshots/nostr-citadel-home-small.png?raw=true "Nostr Citadel Home")

![Startup Screen](screenshots/startup.png?raw=true "Startup Screen")

## Disclaimer
This is me learning [Nostr](https://github.com/fiatjaf/nostr) and [Go](https://go.dev/).  
I will be adding the features that I'd like to see in a sweet relay.

Let me know if there's anything you'd like to see implemented.

No prior Go experience so use it at your own peril. 💀  
It should be [safe](https://www.youtube.com/watch?v=dQw4w9WgXcQ) to use. 

There will be breaking changes up until v0.1.0

As Is - No Warranty!

## Credits
Thanks to [fiatjaf](https://github.com/fiatjaf/relayer) for blueprint and inspiration. 💜

## Features / Todo

- [x] Nostr Relay
    * Core event model
- [x] Npub Whitelist
    * Whitelist your friends and foes for event posting
      * Limit reading in future version for private relays
- [x] Sync data to local relay
    * Sync either all whitelisted npubs or only admin data (notes) to the relay
    * The relay will sync every hour.
      * Using `since` to get new notes only. 
- [x] Simple Cli
  * --port
  * --whitelist -h
  * --invoice -h
- [x] Simple Dashboard
  * Grid of profiles saved on the relay.
  * Vanilla and minimalistic.
  * Disable in config
- [x] Bootstrap Relays
- [x] More Cli
  * More cli things   
- [x] Paid Relay
  * Core Lightning
  * LND (Coming soon)
  * Others... ?
- [ ] Export and Import
  * Export to file
  * Import from file
  * Bootstrap from backup
- [ ] Automatic SSL Termination
  * Lets Encrypt Certificate
- [ ] Citadel Admin & Nostr Client
  * Web - [Nostr Citadel Web Client](https://github.com/MrJohnsson77/nostr-citadel-watch) ( Coming "soon" )
    * Nostr Client & Admin Dashboard for Nostr Citadel
    * Standard client functionality
      * Global
      * Followers
      * Groups
      * Notifications
      * etc...
    * Follow recommendations
    * Trending on your relay ( and others... )
    * AI support, tweak and build your own recommendation engine
      * Export and share your models
- [ ] Additional database backends
  * Redis
  * PostgresSQL
  * MongoDB
  * MySQL
  
## Nips

[NIPs](https://github.com/nostr-protocol/nips) with a relay-specific implementation are listed here.

- [x] NIP-01: [Basic protocol flow description](https://github.com/nostr-protocol/nips/blob/master/01.md)
- [ ] NIP-02: [Contact List and Petnames](https://github.com/nostr-protocol/nips/blob/master/02.md)
- [ ] NIP-03: [OpenTimestamps Attestations for Events](https://github.com/nostr-protocol/nips/blob/master/03.md)
- [ ] NIP-05: [Mapping Nostr keys to DNS-based internet identifiers](https://github.com/nostr-protocol/nips/blob/master/05.md)
- [x] NIP-09: [Event Deletion](https://github.com/nostr-protocol/nips/blob/master/09.md)
- [x] NIP-11: [Relay Information Document](https://github.com/nostr-protocol/nips/blob/master/11.md)
- [x] NIP-11a: Relay Information Document Extensions
- [x] NIP-12: [Generic Tag Queries](https://github.com/nostr-protocol/nips/blob/master/12.md)
- [x] NIP-15: [End of Stored Events Notice](https://github.com/nostr-protocol/nips/blob/master/15.md)
- [x] NIP-16: [Event Treatment](https://github.com/nostr-protocol/nips/blob/master/16.md)
- [x] NIP-20: [Command Results](https://github.com/nostr-protocol/nips/blob/master/20.md)
- [ ] NIP-22: [Event `created_at` limits](https://github.com/nostr-protocol/nips/blob/master/22.md)
- [ ] NIP-26: [Event Delegation](https://github.com/nostr-protocol/nips/blob/master/26.md) 
- [ ] NIP-28: [Public Chat](https://github.com/nostr-protocol/nips/blob/master/28.md)
- [ ] NIP-33: [Parameterized Replaceable Events](https://github.com/nostr-protocol/nips/blob/master/33.md)
- [ ] NIP-42: [Authentication of clients to relays](https://github.com/nostr-protocol/nips/blob/master/42.md)

## Requirements
- Computer
- Internet
- Courage

## Get Started
Download the config.yaml and the binary for your architecture from the [release](https://github.com/MrJohnsson77/nostr-citadel/releases) section.  
Add your npub and relay_url in the config.yaml and drop the config and executable in a folder and run.  

First start will create the database and bootstrap the relay by syncing the admin notes from other relays.
Initial list of relays is downloaded from [nostr.watch](https://api.nostr.watch/v1/online).

A specific bootstrap relay can be set in config.yaml, this relay will be added to the list of relays used during bootstrap.

Open the relay_url in a browser to verify that the relay is running.
Add your relay in your nostr client to connect and start saving your events.

Profiles of whitelisted users will be displayed on the dashboard, it can be disabled in config.yaml by setting `dashboard: false`  
If your profile doesn't show up on the dashboard, start any client and save your profile to push it to your relay.

### Operation
Changing admin_npub in config.yaml will replace the old admin with the new one.  
In this version a change of admin won't purge the events of the old admin, only delete the event with kind 0 (profile)

 *  Bind to specific port
    ```
    ./nostr-citadel start --port 1337 --loglevel INFO
    ```
  
  
  * Display whitelist
    ```
    # List whitelisted npubs  
    
    ./nostr-citadel whitelist --list
    ```

  * Add npub to whitelist
    ```
    ./nostr-citadel whitelist --add npub....
    ```

  * Remove npub from whitelist
    ```
    # Removing a npub from the whitelist will delete all events for it too.  
    
    ./nostr-citadel whitelist --remove npub....
    ```

  * Create invoice for npub
    ```
    ./nostr-citadel invoice --create npub....
    ```
    
  * Create QR Code invoice for npub
    ```
    ./nostr-citadel invoice --qr npub....
    ```
    
  * Verify if invoice is paid
    ```
    ./nostr-citadel invoice --verify npub....
    ```

### Run from binary
Download binary for your architecture from the [releases](https://github.com/MrJohnsson77/nostr-citadel/releases) section.

### Run from source
  ```
  $ git clone git@github.com:MrJohnsson77/nostr-citadel.git
  $ cd nostr-citadel
  $ go run main
  ```

### Build from source
  ```
  $ git clone git@github.com:MrJohnsson77/nostr-citadel.git
  $ cd nostr-citadel
  $ make build
  ```

# Author

- [MrJohnsson](https://github.com/MrJohnsson77) - npub1fhdx6c6pt0ff6k3h5em760fzzzcehe9kqnjl05d2xwmg0ctjp80sn4hhsv 


# Contributors (A-Z)

- 

## License

This is free and unencumbered software released into the public domain.  
For more information, please refer to <http://unlicense.org/>