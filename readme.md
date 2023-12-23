# Eureka
### Web tool for searching a phrase in given set of documents

+ Downloads the document and search for the phrase queried
+ Returns the most relevant documents, along with the snippets including the given query phrase

+ App is divided mainly into two services:
  + Eureka Doorman API: This is responsible for the interaction with the user. Receiving the set of links for documents and returning the search result.
  + Eureka Librarian API: This is responsible for downloading the documents, search for the phrase and return relevant documents to the Doorman API. Three instances of this service is used.

<img src="https://drive.google.com/file/d/1kujllrPYtJPsq3Hv03DXCd1OEXbAVeXK/view?usp=drive_link">

+ The users will interact with the app through /api/feeder and /api/query APIs
+ The Doorman server interact with Librarian servers through /api/index and /api/query APIs 