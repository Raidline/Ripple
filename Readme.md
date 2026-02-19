# Ripple

## Next Steps

- [] Put live watch of file in TUI warning of possible impact in other files
- [] Create local LLM to learn with the graph and be able to make complex queries

## Improvements

- [] Remove tests files (specific per language?, configurable by project)
- [] Start script is not clear what params we are sending, make it more clear
- [] Put the graph in a file (to serve as persistent storage) as rebuild like that to now creep the entire project again
- [] Always run the graph building in background and compare to what we have (new files could have been added while the tool was not running)
- [] Act on file create and rename to update the graph (for now we need to run the project again)

## Optional

There has been the idea to make this a Intellij plugin. 
When everything done, see if that is feasible, calling this tool in kotlin code or finding another way.