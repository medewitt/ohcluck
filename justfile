all:
  letitrip; render; serve;

# Compile letitrip
letitrip:
  cd letitrip; just compile;

# Render
render:
  quarto render;

# Serve locally
serve:
  quarto preview;
