document.addEventListener('DOMContentLoaded', () => {
  const columnDefinitions = [
    { headerName: 'ID', field: 'ID', editable: true, flex: 0.5 },
    { headerName: 'Name', field: 'Name', filter: true, editable: true, flex: 1.5 },
    {
      headerName: 'Datum', field: 'Date', editable: true, flex: 1, cellDataType: 'text',
      comparator: (valueA, valueB) => {
        const parse = (dateStr) => {
          if (!dateStr) return 0;
          const [day, month, year] = dateStr.split('.').map(Number);
          return new Date(year, month - 1, day).getTime();
        };
        return parse(valueA) - parse(valueB);
      }
    },
    { headerName: 'Ort', field: 'Place', filter: true, editable: true, flex: 1 },
    { headerName: 'Stadt', field: 'City', filter: true, editable: true, flex: 1 },
    { headerName: 'Link', field: 'Link', editable: true, flex: 1, }
  ];

  const options = {
    columnDefs: columnDefinitions,
    rowSelection: {
      mode: 'multiRow'
    },
    onCellValueChanged: async (params) => {
      if (params.oldValue === params.newValue) return;
      try {
        const body = new URLSearchParams({
          id: params.data.ID,
          field: params.colDef.field,
          value: params.newValue,
        });

        fetch('/events/manage/api/change', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body
        });
      } catch (err) {
        console.log(err);
      }
    }
  };

  // Getting data from
  fetch('/events/manage/api/load')
    .then(response => response.json())
    .then(data => {
      gridApi.setGridOption("rowData", data);
    })
    .catch(error => {
      // TODO: When network fails, this data could be rendered instead
      // mock = [{"ID":22,"Name":"Kunstkomplex","Date":"30.08.2025","Place":"Theater K","City":"Oldenburg","Link":null},{"ID":21,"Name":"Poetry Slam","Date":"27.09.2025","Place":"Schloss Fechenbach","City":"Dieburg","Link":null},{"ID":20,"Name":"7:77","Date":"16.09.2025","Place":"Neues Schauspiel","City":"Leipzig","Link":"https://neuesschauspielleipzig.de/spielplan/7-77-die-open-stage/date:1758045600"},{"ID":19,"Name":"Pinzette und Kneifzange (Lesebühne)","Date":"11.09.2025","Place":"Bayerhaus","City":"Leipzig","Link":null},{"ID":18,"Name":"Lila Comedy","Date":"10.09.2025","Place":"Radtanke","City":"Leipzig","Link":null},{"ID":17,"Name":"Poetry\u0026Hip-Hop","Date":"25.10.2025","Place":"Moritzbastei","City":"Leipzig","Link":"https://www.moritzbastei.de/event/konzert/2025-10-25/poetry-hip-hop-2025"},{"ID":16,"Name":"Einfach Elses Late Night Show","Date":"15.10.2025","Place":"Moritzbastei","City":"Leipzig","Link":"https://www.moritzbastei.de/event/show/2025-10-15/einfach-elses-late-night-show#:~:text=Mi.,.2025%2C%2020%3A00%20Uhr\u0026text=Die%20erste%20und%20letzte%20Late,Es%20wird%20deep"},{"ID":15,"Name":"Queer Slam","Date":"14.10.2025","Place":"Werk 2","City":"Leipzig","Link":"https://www.tixforgigs.com/de-DE/Event/66821/queer-slam-leipzig-14-10-2025-werk-2-kulturfabrik-leipzig-e-v-halle-a"},{"ID":14,"Name":"Safer Space Slam","Date":"19.11.2025","Place":"AStA Uni","City":"Hamburg","Link":"https://saferspace-slam.de/"},{"ID":13,"Name":"Moderation Podiumsdiskussion: Poetry Slam - Jugendlyrik oder Weltliteratur?","Date":"08.11.2025","Place":"Universitäts Bibliothek Chemnitz","City":"Chemnitz","Link":null},{"ID":12,"Name":"Weihnachtsfeier","Date":"18.12.2025","Place":"HArfE e.V.","City":"Hameln","Link":null},{"ID":11,"Name":"Bamberg ist Slamberg","Date":"15.12.2025","Place":"Haas Säle","City":"Bamberg","Link":"https://vivenu.com/event/poetry-slam-bamberg-ist-slamberg-d9ie4l"},{"ID":10,"Name":"Poetry Slam Würzburg - Frühe Bescherung","Date":"14.12.2025","Place":"Posthalle","City":"Würzburg","Link":"https://www.eventbrite.de/e/poetry-slam-wurzburg-fruhe-bescherung-tickets-1480593046659"},{"ID":9,"Name":"Poetry Slam","Date":"13.12.2025","Place":"Moritzhof","City":"Magdeburg","Link":"https://moritzhof-magdeburg.de/live/"},{"ID":8,"Name":"Queer Slam","Date":"17.07.2026","Place":"tba","City":"Dortmund","Link":null},{"ID":7,"Name":"Queer Slam","Date":"11.06.2026","Place":"Moritzbastei","City":"Leipzig","Link":null},{"ID":6,"Name":"Queer Slam","Date":"08.04.2026","Place":"tba","City":"Bochum","Link":null},{"ID":5,"Name":"Einfach Elses Late Night Show","Date":"11.02.2026","Place":"Moritzbastei","City":"Leipzig","Link":"https://www.tixforgigs.com/de-de/Event/68817"},{"ID":4,"Name":"Poetry and HipHop","Date":"04.02.2026","Place":"Moritzbastei","City":"Leipzig","Link":null},{"ID":3,"Name":"Einfach Elses Late Night Show","Date":"30.01.2026","Place":"Atomino","City":"Chemnitz","Link":"https://www.tixforgigs.com/de-DE/Event/69963/einfach-elses-late-night-show-atomino-chemnitz"},{"ID":2,"Name":"7:77","Date":"20.01.2026","Place":"Neues Schauspielhaus","City":"Leipzig","Link":"https://openstage777.com/termine/"},{"ID":1,"Name":"Westslam","Date":"15.01.2026","Place":"Neues Schauspielhaus","City":"Leipzig","Link":"https://rausgegangen.de/en/events/westslam-17/"}]
      console.error('Error:', error);
      gridDiv.innerHTML = `<div style="color: red;">Error: ${error.message}</div>`;
    });

  const gridDiv = document.getElementById('events-grid');
  gridApi = agGrid.createGrid(gridDiv, options);
});

document.getElementById('event-create-form').addEventListener('submit', async (e) => {
  e.preventDefault(); // do NOT reload the page

  const form = e.target;

  // const formData = new FormData(form);
  // console.log(Object.fromEntries(formData));

  const body = new URLSearchParams(new FormData(form));
  fetch('/events/manage/api/create', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body
  });

  // Clear the form when a submit has happend
  setTimeout(() => {
    e.target.reset();
  }, 0);

  // Reload the agGrid
  try {
    const res = await fetch('/events/manage/api/load');
    const data = await res.json();
    gridApi.setGridOption("rowData", data);
  } catch (err) {
    console.error('Reload failed:', err);
  }
});

