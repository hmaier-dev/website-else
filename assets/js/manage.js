document.addEventListener('DOMContentLoaded', () => {
  let rowData = [
    {"Name":"Poetry Slam \u0026 Moderation","Date":"2026-07-31T00:00:00Z","Doors":"2026-07-31T00:00:00Z","Begin":"2026-07-31T00:00:00Z","Place":"Crossover Festival","City":"Grimma","Coordinates":"","Link":null,"Hash":"9XkewXXag94E9CTQfRndWNWiEZQjLv3BznxHyyjdKupb","Separator":null,"Upcomming":null},
    {"Name":"Queer Slam","Date":"2026-07-17T00:00:00Z","Doors":"2026-07-17T00:00:00Z","Begin":"2026-07-17T00:00:00Z","Place":"tba","City":"Dortmund","Coordinates":"","Link":null,"Hash":"9ej32pQDyDX6r7qJizkGeGJw9Yy1KAxQwkFTeHyqdzTV","Separator":null,"Upcomming":null},
    {"Name":"Moderation Jugend Journalismus Preis","Date":"2026-06-20T00:00:00Z","Doors":"2026-06-20T00:00:00Z","Begin":"2026-06-20T00:00:00Z","Place":"tba","City":"Dresden","Coordinates":"","Link":null,"Hash":"2QEjKocEu63i4JA1Ez5SoA29kV3CEyvyfRs3q5s7tEfj","Separator":null,"Upcomming":null},
    {"Name":"Queer Slam Open Air","Date":"2026-06-11T00:00:00Z","Doors":"2026-06-11T00:00:00Z","Begin":"2026-06-11T00:00:00Z","Place":"Moritzbastei","City":"Leipzig","Coordinates":"","Link":null,"Hash":"CyqdJQ1zzJzbhNxVp5XdUHJfriMN5BZyWRbKH9PVHrgb","Separator":null,"Upcomming":null},
    {"Name":"Auftritt Poetry \u0026 Co.","Date":"2026-06-05T00:00:00Z","Doors":"2026-06-05T00:00:00Z","Begin":"2026-06-05T00:00:00Z","Place":"tba","City":"Falkenlandkreis","Coordinates":"","Link":null,"Hash":"7gh2xLYLC4g6xanMec6vURdVgb85PoV7FMHw9CJ3Vqf4","Separator":null,"Upcomming":null},
    {"Name":"Einfach Elses Late Night Show","Date":"2026-05-06T00:00:00Z","Doors":"2026-05-06T00:00:00Z","Begin":"2026-05-06T00:00:00Z","Place":"Moritzbastei","City":"Leipzig","Coordinates":"","Link":null,"Hash":"6V9Xf19gKA8nnWmyMPyGVvqiWTpqXjGAz4hB3ZFNDGSx","Separator":null,"Upcomming":null},
    {"Name":"Einfach Elses Late Night Show","Date":"2026-04-24T00:00:00Z","Doors":"2026-04-24T00:00:00Z","Begin":"2026-04-24T00:00:00Z","Place":"Atomino","City":"Chemnitz","Coordinates":"","Link":null,"Hash":"6hkdR9BFYXx2ARuqsJhCWLUrJ3UTqYxs53pXa6yaerLS","Separator":null,"Upcomming":null},
    {"Name":"Queer Slam","Date":"2026-04-22T00:00:00Z","Doors":"2026-04-22T00:00:00Z","Begin":"2026-04-22T00:00:00Z","Place":"tba","City":"Leipzig","Coordinates":"","Link":null,"Hash":"AMRby4rT5zBCwjktakszKQ2AHArv247LoRis2HWb6y2e","Separator":null,"Upcomming":null},
    {"Name":"Queer Slam","Date":"2026-04-22T00:00:00Z","Doors":"2026-04-22T00:00:00Z","Begin":"2026-04-22T00:00:00Z","Place":"Werk 2 (Halle D)","City":"Leipzig","Coordinates":"","Link":"https://www.tixforgigs.com/de-DE/Event/72065/queer-slam-leipzig-april-werk-2-kulturfabrik-leipzig-e-v-halle-d","Hash":"E2oSS65vjkX1BfHfsDof6ciaqQEUrSUapjVi8Fjww7aa","Separator":null,"Upcomming":null},
    {"Name":"Queer Slam","Date":"2026-07-17T00:00:00Z","Doors":"2026-07-17T00:00:00Z","Begin":"2026-07-17T00:00:00Z","Place":"tba","City":"Dortmund","Coordinates":"","Link":null,"Hash":"J2cGXEvkgodWyC4Xq5M8DkGvmLDiFZXtG1NSSkpdm9mu","Separator":null,"Upcomming":null}
  ];
   
  const columnDefs = [
    { headerName: 'Name', field: 'Name', editable: true, flex: 1.5 },
    {
      headerName: 'Datum',
      field: 'Date',
      editable: true,
      flex: 1,
      cellDataType: 'dateString',
      valueFormatter: (params) => {
        if (!params.value) return '';
        const date = new Date(params.value);
        const day = String(date.getDate()).padStart(2, '0');
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const year = date.getFullYear();
        return `${day}.${month}.${year}`;
      },
      valueParser: (params) => {
        if (!params.newValue) return null;
        const [day, month, year] = params.newValue.split('.');
        return `${year}-${month}-${day}`;
      }
    },
    { headerName: 'Ort', field: 'Place', editable: true, flex: 1 },
    { headerName: 'Stadt', field: 'City', editable: true, flex: 1 },
    { headerName: 'Link', field: 'Link', editable: true, flex: 1 }
  ];


  const gridOptions = {
    columnDefs: columnDefs,
    rowData: rowData,
    rowSelection: {
    mode: 'multiRow'
    }
    };

  // Initialize the grid
  const gridDiv = document.getElementById('events-grid');
  gridApi = agGrid.createGrid(gridDiv, gridOptions);
  

});
