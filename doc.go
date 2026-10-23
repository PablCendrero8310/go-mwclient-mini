/*
Package mwclientmini provides functionality for basic interacting with the MediaWiki API

It is intended for people who knows the API and wants a way to make easier the hard parts.

# Basic usage

In the example below, basic usage of go-mwclient-mini is shown.

	// Init a *MediaWikiClient with New() and specifying the wiki's API URL, the user-agent, and the login data

	client := mwclientmini.New()

	params := url.Values{
				"action":   "query",
				"list":     "recentchanges",
		}
		response, err := wiki.Get(params)
		if err != nil {
			panic(err)
		}

go-mwclient-mini only offers Get method for now, but I am planning to add others like Post.

It also offers login method only with BotPasswords, but planning to add OAuth too.

# Error handling

If an API call fails it will return an error. Many things can go wrong
during an API call: the network could be down, the API could return an
unexpected response (if the API was changed), or perhaps there's an
error in your API request.

If the error is an API error or warning, then the error/warning(s) will be returned as error

For more information about API errors and warnings, please see
https://www.mediawiki.org/wiki/API:Errors_and_warnings.
*/
package mwclientmini
