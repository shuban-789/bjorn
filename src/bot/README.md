# Bot Notes

Quick notes on how a few things work:
1. All components unique ids are named based on the following convention: `commandname;componentname componentdata`
  - EX: `team;awards_n %d_%d_%d`
  - This is essential for how handlers are called: The component data segment is not matched, but the commandname and componentname are (so commandname;componentname should be unique)
  - Use the generateComponentId() function for this
  - Remember that in total it must be less than 100 charas, so use abbreviations where possible
2. 