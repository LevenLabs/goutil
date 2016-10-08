# stdin-llog

A simple binary for outputting results from stdin in [llog](https://github.com/levenlabs/go-llog) format.

## Examples
```
> echo "2" | stdin-llog 1+1=
~ INFO -- 1+1= -- value="2"
```

```
> echo "2" | stdin-llog --key=answer 1+1=
~ INFO -- 1+1= -- answer="2"
```

```
> echo "h
> e
> l
> l
> o" | stdin-llog hello
~ INFO -- hello -- value="h"
~ INFO -- hello -- value="e"
~ INFO -- hello -- value="l"
~ INFO -- hello -- value="l"
~ INFO -- hello -- value="o"
```
