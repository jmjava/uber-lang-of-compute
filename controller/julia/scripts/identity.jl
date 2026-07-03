#!/usr/bin/env julia
# identity: pass input JSON through unchanged.
input_path, output_path = ARGS[1], ARGS[2]
write(output_path, read(input_path, String))
