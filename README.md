# mtg-cr-tool

A simple Go application that converts the _Magic: The Gathering_ comprehensive rules into Markdown files ready for Retrieval Augmented Generation (RAG).

## Usage

mtg-cr-tool requires the current ruleset in plain text.
An up-to-date link can be found on the official [_Magic: The Gathering_ rules website](https://magic.wizards.com/en/rules).

```bash
# Fetch comprehensive rules (this URL may be out of date)
curl --fail -o rules.txt https://media.wizards.com/2026/downloads/MagicCompRules%2020260819.txt
```

Pass the downloaded file to the application along with an output directory.
If the directory doesn't exist, it will be created.

```
go run . -cr rules.txt -o ./output
```

## License

Apache 2.0
