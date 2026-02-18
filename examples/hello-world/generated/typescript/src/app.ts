// Auto-generated TypeScript/Express server from GlyphLang
// Do not edit manually

import express, { Request, Response } from 'express';


const app = express();
app.use(express.json());

app.get('/', async (req: Request, res: Response) => {
  return res.json({ text: "Hello, World!", timestamp: 1234567890 });
});

app.get('/hello/:name', async (req: Request, res: Response) => {
  const name = req.params.name;
  const greeting = (("Hello, " + name) + "!");
  return res.json({ message: greeting });
});

app.get('/health', async (req: Request, res: Response) => {
  return res.json({ status: "ok" });
});

app.listen(3000, '0.0.0.0', () => {
  console.log(`Server running on http://0.0.0.0:3000`);
});
