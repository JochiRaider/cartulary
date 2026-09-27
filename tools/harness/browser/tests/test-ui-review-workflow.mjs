import test from "node:test";
import { publicWorkflow } from "./ui-review-public-workflow.mjs";

test("public dev and browser-free artifact workflows keep private detail out of retained receipts and expire all links", async () => { await publicWorkflow(); });
