# Bot Notes

Quick notes on how a few things work:
1. All components unique ids are named based on the following convention: `commandname;componentname componentdata`
  - EX: `team;awards_n %d_%d_%d`
  - This is essential for how handlers are called: The component data segment is not matched, but the commandname and componentname are (so commandname;componentname should be unique)
  - Use the generateComponentId() function for this
  - Remember that in total it must be less than 100 charas, so use abbreviations where possible


#### regions.csv
Note: regions.csv is pulled from [regions.ts](https://github.com/ftc-scout/ftc-scout/blob/2059f795a78cc7b091189dbb493444b90d91f236/packages/web/src/lib/util/regions.ts#L19) in the FTCScout GitHub

You will have to update the bot each time a new region is added, but realistically I highly doubt anyoone from one of these new regions will be using Bjorn. As far as I am aware, there is no API to pull the list of existing regions. Perhaps the FTC Events API has this? idk