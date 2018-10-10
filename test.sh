#!/bin/bash
./fractal -exit=true -seg=1 -output=00.png #-x=-0.661 -y=-0.35 -z=1000000 -i=3000
./fractal -exit=true -seg=2 -output=10.png #-x=-0.661 -y=-0.35 -z=1000000 -i=3000
./fractal -exit=true -seg=3 -output=20.png #-x=-0.661 -y=-0.35 -z=1000000 -i=3000
./fractal -exit=true -seg=4 -output=01.png #-x=-0.661 -y=-0.35 -z=1000000 -i=3000
./fractal -exit=true -seg=5 -output=11.png #-x=-0.661 -y=-0.35 -z=1000000 -i=3000
./fractal -exit=true -seg=6 -output=21.png #-x=-0.661 -y=-0.35 -z=1000000 -i=3000
convert 00.png 10.png 20.png +append top.png
convert 01.png 11.png 21.png +append bottom.png
convert top.png bottom.png -append fractal.png
rm 00.png 10.png 20.png 01.png 11.png 21.png top.png bottom.png