function draw_grid(canvas, size) {

  var width = canvas.width;
  var height = canvas.height;

  var j = 0;
  var line = null;
  var rect = [];

  console.log(width + ":" + height);

  for (var i = 0; i < Math.ceil(width / size); ++i) {
    rect[0] = i * size;
    rect[1] = 0;

    rect[2] = i * size;
    rect[3] = height;

    line = null;
    line = new fabric.Line(rect, {
      stroke: '#999',
      opacity: 0.5,
    });

    line.selectable = false;
    canvas.add(line);
    line.sendToBack();

  }

  for (i = 0; i < Math.ceil(height / size); ++i) {
    rect[0] = 0;
    rect[1] = i * size;

    rect[2] = width;
    rect[3] = i * size;

    line = null;
    line = new fabric.Line(rect, {
      stroke: '#999',
      opacity: 0.5,
    });
    line.selectable = false;
    canvas.add(line);
    line.sendToBack();

  }

}

function mm2pix(mm) {
  return m2pix(mm / 1000);
}

function m2pix(m) {
  return m * 5
}

function pix2mm(pix) {
  return pix * 1000 / 5;
}

function fabricInit() {
  var canvas = new fabric.Canvas('c');
  canvas.setHeight(700);
  canvas.setWidth(1000);
  return canvas;
}

function render(canvas, data) {
  canvas.clear();
  draw_grid(canvas, m2pix(10)); // 10 meters per cell

  var rainbow = ["#ffcc00", "#ccff00", "#00ccff", "#ff0000", "#ffff00"];
  for (var i=0; i < data.length; i++) {
    var text = new fabric.Text(String(data[i].I), {fontSize: 16, fill: 'black'});
    var circle = new fabric.Circle({radius: 10, fill: rainbow[i % rainbow.length]});
    var group = new fabric.Group([circle, text], {
      left: mm2pix(data[i].X), top: mm2pix(data[i].Y)
    });
    group.nodeData=data[i]

    canvas.add(group);
  }

  var canvasOnChange = function(options) {
    options.target.setCoords();
    canvas.forEachObject(function(obj) {
      if (obj === options.target) return;
      obj.setOpacity(options.target.intersectsWithObject(obj) ? 0.5 : 1);
      node = options.target.nodeData
      node.X = pix2mm(options.target.left);
      node.Y = pix2mm(options.target.top);
      $.post('set', JSON.stringify(node));
    });
  }

  canvas.on({
    'object:moving': canvasOnChange,
    'object:scaling': canvasOnChange,
    'object:rotating': canvasOnChange,
  });

  canvas.renderAll();
}


function fetchData() {
  $.getJSON('list', function(data){
    render(fabricInit(), data);
  });
}

$(window).load(fetchData);
