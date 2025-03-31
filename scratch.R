library(rvest)

url <- "https://www.thelancet.com/journals/lanres/article/PIIS2213-2600(25)00052-9/fulltext"

ses <- read_html(url)

