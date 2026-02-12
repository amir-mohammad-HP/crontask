const express = require("express");
const path = require("path");
const fs = require("fs").promises;
const bodyParser = require("body-parser");
const methodOverride = require("method-override");
const expressLayouts = require("express-ejs-layouts");

const app = express();
const PORT = process.env.PORT || 3000;

// Set up EJS as templating engine
app.set("view engine", "ejs");
app.set("views", path.join(__dirname, "views"));

// Use express-ejs-layouts
app.use(expressLayouts);
app.set("layout", "layout");
app.set("layout extractScripts", false);
app.set("layout extractStyles", false);

// Middleware
app.use(express.static(path.join(__dirname, "public")));
app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());
app.use(methodOverride("_method"));

// Ensure directories exist
const diagramsDir = path.join(__dirname, "data", "diagrams");
const mermaidDir = path.join(__dirname, "data", "mermaid");
fs.mkdir(diagramsDir, { recursive: true }).catch(console.error);
fs.mkdir(mermaidDir, { recursive: true }).catch(console.error);

// Helper function to convert title to filename
function titleToFilename(title) {
  return (
    title
      .replace(/\//g, "-") // Replace / with -
      .replace(/\\/g, "-") // Replace \ with -
      .replace(/\./g, "-") // Replace . with -
      .toLowerCase() + // Convert to lowercase
    ".mmd"
  );
}

// Helper functions
async function getAllDiagrams() {
  try {
    const files = await fs.readdir(diagramsDir);
    const jsonFiles = files.filter((file) => file.endsWith(".json"));

    const diagrams = await Promise.all(
      jsonFiles.map(async (file) => {
        const content = await fs.readFile(path.join(diagramsDir, file), "utf8");
        const diagram = JSON.parse(content);

        // Read mermaid code from .mmd file
        const mermaidFilename = titleToFilename(diagram.title);
        const mermaidPath = path.join(mermaidDir, mermaidFilename);

        try {
          const mermaidCode = await fs.readFile(mermaidPath, "utf8");
          diagram.mermaidCode = mermaidCode;
        } catch (err) {
          console.error(
            `Mermaid file not found for ${diagram.title}: ${mermaidFilename}`,
          );
          diagram.mermaidCode =
            "flowchart TD\n    Start([Error]) --> End([Missing Diagram File])";
        }

        return diagram;
      }),
    );

    return diagrams.sort((a, b) => a.id - b.id);
  } catch (error) {
    console.error("Error reading diagrams:", error);
    return [];
  }
}

async function getDiagramById(id) {
  try {
    const filePath = path.join(diagramsDir, `${id}.json`);
    const content = await fs.readFile(filePath, "utf8");
    const diagram = JSON.parse(content);

    // Read mermaid code from .mmd file
    const mermaidFilename = titleToFilename(diagram.title);
    const mermaidPath = path.join(mermaidDir, mermaidFilename);

    try {
      const mermaidCode = await fs.readFile(mermaidPath, "utf8");
      diagram.mermaidCode = mermaidCode;
    } catch (err) {
      console.error(
        `Mermaid file not found for ${diagram.title}: ${mermaidFilename}`,
      );
      diagram.mermaidCode =
        "flowchart TD\n    Start([Error]) --> End([Missing Diagram File])";
    }

    return diagram;
  } catch (error) {
    return null;
  }
}

// Routes
app.get("/", async (req, res) => {
  try {
    const diagrams = await getAllDiagrams();
    res.render("index", {
      layout: "layout",
      title: "Go Application Execution Flow",
      diagrams,
      currentYear: new Date().getFullYear(),
      lastUpdated: new Date().toLocaleDateString(),
    });
  } catch (error) {
    console.error("Error loading diagrams:", error);
    res.status(500).send("Error loading diagrams");
  }
});

app.get("/diagram/:id", async (req, res) => {
  try {
    const diagram = await getDiagramById(req.params.id);
    if (!diagram) {
      return res.status(404).send("Diagram not found");
    }
    res.render("diagram", {
      layout: "layout",
      title: `${diagram.title} - Diagram Details`,
      diagram,
      currentYear: new Date().getFullYear(),
    });
  } catch (error) {
    console.error("Error loading diagram:", error);
    res.status(500).send("Error loading diagram");
  }
});

app.get("/api/diagrams", async (req, res) => {
  try {
    const diagrams = await getAllDiagrams();
    res.json(diagrams);
  } catch (error) {
    res.status(500).json({ error: "Error fetching diagrams" });
  }
});

app.get("/api/diagrams/:id", async (req, res) => {
  try {
    const diagram = await getDiagramById(req.params.id);
    if (!diagram) {
      return res.status(404).json({ error: "Diagram not found" });
    }
    res.json(diagram);
  } catch (error) {
    res.status(500).json({ error: "Error fetching diagram" });
  }
});

// Start server
app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});
