# Walkthrough

Hey there! In this document, we'll use the [PokéAPI](https://pokeapi.co/) to exercise some of the utilities found in `sqump`, in an effort to demonstrate what's possible with the tools available.

In order to follow along with this tutorial, it's assumed that you have `sqump` installed. For installation instructions, refer back to the `README.md` in the base of the project.

## Setup
With `sqump` installed, we'll go ahead and run `sqump init`. This should do two things. First, it will prompt us to create our initial core config, which we'll confirm. Second, it will create a `Squmpfile.json` in the current directory, which we will use to build our scripts.

Next, we can use two paths to start developing: the CLI or the web UI. To go with the web UI, run `sqump webview`, which should print a local URL at which the UI will be availble. To use the CLI, we'll want to set the `EDITOR` variable in our environment, to facilitate some of the interactive development features. I personally use [Neovim](https://neovim.io/), so I put the following in my `.bashrc`: `export EDITOR='nvim'`. Now we should be good to go! For the rest of the guide, we'll try to include instructions for both the web UI and CLI options.

## Writing our First Script
Now we can get down to it! Open your web UI to the new request that was created, or enter `sqump edit` to fuzzy-find that same request. There should also be options to edit the collection and request name in both instances (web UI option coming soon).

Once you're editing your request, we'll enter this code:
```lua
local s = require('sqump')

local resp = s.fetch('https://pokeapi.co/api/v2/pokemon/{{.pokemon}}')

s.print_response(resp)
```

This code will be used to retrieve information about a particular Pokémon and display it.

### Aside: Environments
You may have noticed the odd `{{.pokemon}}` at the end of the `fetch` URL. This is a template placeholder that pulls data from the configured environment. In the UI, you should be able to see the current environment on the request page, along with the placeholder fields it was created with. To see the current environment via CLI, use `sqump info core`.

The initial value for the current environment should be `staging`, so we'll add the field `"pokemon": "pikachu"` to this environment. This should be possible via the Environment panel on the web UI request page, or via `sqump edit` on the CLI.

As a final (and important!) note, there are three scopes of environment variable provided for your convenience - `collection`, `core`, and `temporary`. In order of precedence, temporary overrides core which overrides collection. This is arranged so that sensible defaults can be included with a script, but the end user is allowed to override these values on a temporary or relatively permanent basis.

With our environment explained and configured, we can get back to writing our scripts!

### Back to Our First Script
Now if we run the script as it exists, we should see a rather large print looking like the following:
```
Status Code: 200

Headers:
Cf-Ray: [84b2b174ae0d4788-DFW]

... truncated for clarity ...

 "types": [
    {
      "slot": 1,
      "type": {
        "name": "electric",
        "url": "https://pokeapi.co/api/v2/type/13/"
      }
    }
  ],
  "weight": 60
}
```
If so, we've successfully retrieved our information about Pikachu! Now, any time we want to run this against a different Pokémon, we can simply change the environment variable.

Now that we've seen the script roughly working, we can explore some other capabilities `sqump` offers. Let's replace our script with the below:
```lua
local s = require('sqump')

local resp = s.fetch('https://pokeapi.co/api/v2/pokemon/{{.pokemon}}')

print(s.drill_json('id', resp.body), s.drill_json('types.1.type.name', resp.body))

s.export({
  pokemon = resp.body,
})
```
With this, we can check Pikachu's Pokédex ID and the name of his typing, then export the data we gathered to be used by another script! Heads up for the 1-based Lua indexing in `drill_json`. Next, we'll look at starting to use this information in another script.

### Bringing in Kafka
For this next portion, we'll assume that you have Docker installed. We'll be using it to run the `docker-compose.yml` file in this directory to bring up a local Kafka instance to test with.

Let's go ahead and spin up that Kafka instance by executing `docker compose -f docker-compose.yml up` with that file, and waiting for it to finish coming up.

After that's done, we'll go ahead and create a new script with the following:
```lua
local s = require('sqump')

local res = s.execute('Req1')

local weight = s.drill_json('weight', res.pokemon)
local id = s.drill_json('id', res.pokemon)

local k = require('sqump_kafka')
local brokers = {'localhost:9092'}
local topic = 'weights'

k.provision_topic(brokers, topic)

local p = k.new_producer(brokers, topic, 5)
p:write(tostring(id), tostring(weight))
p:close()
print('done writing!')
```
Here, we're taking the results from our Pokémon search script and writing the weight to a new topic in our Kafka cluster. We're even making sure to provision the topic to ensure it exists before writing!

Next, let's setup up a quick script to consume from our new topic and average the weights we receive:
```lua
local k = require('sqump_kafka')

local brokers = {'localhost:9092'}
local topic = 'weights'

math.randomseed(os.time())
local c = k.new_consumer(brokers, string.format("weights-%d", math.random(1, 1000)), topic)

local sum, count = 0, 0
for i=1,2 do
  local msg = c:read_message(15)
  sum = sum + msg.data
  count = count + 1
end
c:close()

print('Total weight:', sum)
print('Average:', sum / count)
```
Here we create a consumer with a randomized group ID to get the full run each time, then iterate over the messages to sum and then average the weights.

That's all for this initial walkthrough, to show the basic options that are available! Be sure to check out the API docs in this directory to see what all else is available.
