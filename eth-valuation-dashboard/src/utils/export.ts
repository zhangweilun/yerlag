/**
 * Data export utilities for the ETH Valuation Dashboard.
 * Supports CSV data export and chart export as PNG/SVG.
 */

/**
 * Escapes a CSV field value by wrapping in quotes if it contains
 * commas, quotes, or newlines. Internal quotes are doubled.
 */
function escapeCSVField(value: string): string {
  if (value.includes(',') || value.includes('"') || value.includes('\n') || value.includes('\r')) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

/**
 * Exports tabular data as a CSV file download.
 * Handles proper escaping of commas and quotes, adds BOM for Excel compatibility,
 * and uses CRLF line endings per RFC 4180.
 */
export function exportToCSV(
  headers: string[],
  records: Record<string, unknown>[],
  filename: string
): void {
  const csvContent = generateCSVContent(headers, records);

  // Add BOM for Excel compatibility with UTF-8
  const BOM = '\uFEFF';
  const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8;' });

  triggerDownload(blob, filename.endsWith('.csv') ? filename : `${filename}.csv`);
}

/**
 * Generates CSV string content from headers and records.
 * Exported separately for testability.
 */
export function generateCSVContent(
  headers: string[],
  records: Record<string, unknown>[]
): string {
  const headerLine = headers.map(escapeCSVField).join(',');

  const dataLines = records.map((record) => {
    return headers
      .map((header) => {
        const value = record[header];
        if (value === null || value === undefined) {
          return '';
        }
        return escapeCSVField(String(value));
      })
      .join(',');
  });

  // Use CRLF line endings per RFC 4180
  // Add trailing CRLF to ensure round-trip correctness for records with empty values
  return [headerLine, ...dataLines].join('\r\n') + '\r\n';
}

/**
 * Parses CSV content back into structured data.
 * Handles quoted fields with escaped quotes and CRLF/LF line endings.
 */
export function parseCSV(csvContent: string): { headers: string[]; records: Record<string, string>[] } {
  // Remove BOM if present
  const content = csvContent.startsWith('\uFEFF') ? csvContent.slice(1) : csvContent;

  const lines = parseCSVLines(content);
  if (lines.length === 0) {
    return { headers: [], records: [] };
  }

  const headers = lines[0];
  const records: Record<string, string>[] = [];

  for (let i = 1; i < lines.length; i++) {
    const fields = lines[i];
    const record: Record<string, string> = {};
    for (let j = 0; j < headers.length; j++) {
      record[headers[j]] = fields[j] ?? '';
    }
    records.push(record);
  }

  return { headers, records };
}

/**
 * Parses CSV content into an array of field arrays (one per line).
 * Handles quoted fields containing commas, newlines, and escaped quotes.
 */
function parseCSVLines(content: string): string[][] {
  const lines: string[][] = [];
  let currentFields: string[] = [];
  let currentField = '';
  let inQuotes = false;
  let i = 0;

  while (i < content.length) {
    const char = content[i];

    if (inQuotes) {
      if (char === '"') {
        // Check for escaped quote (double quote)
        if (i + 1 < content.length && content[i + 1] === '"') {
          currentField += '"';
          i += 2;
        } else {
          // End of quoted field
          inQuotes = false;
          i++;
        }
      } else {
        currentField += char;
        i++;
      }
    } else {
      if (char === '"') {
        inQuotes = true;
        i++;
      } else if (char === ',') {
        currentFields.push(currentField);
        currentField = '';
        i++;
      } else if (char === '\r') {
        // Handle CRLF
        currentFields.push(currentField);
        currentField = '';
        lines.push(currentFields);
        currentFields = [];
        i++;
        if (i < content.length && content[i] === '\n') {
          i++;
        }
      } else if (char === '\n') {
        currentFields.push(currentField);
        currentField = '';
        lines.push(currentFields);
        currentFields = [];
        i++;
      } else {
        currentField += char;
        i++;
      }
    }
  }

  // Push the last field and line if there's remaining content
  if (currentField !== '' || currentFields.length > 0) {
    currentFields.push(currentField);
    lines.push(currentFields);
  }

  return lines;
}

/**
 * Exports a chart element as a PNG image.
 * Uses the Canvas API to render the element content.
 */
export function exportChartAsPNG(chartRef: HTMLElement, filename: string): void {
  const canvas = document.createElement('canvas');
  const rect = chartRef.getBoundingClientRect();
  canvas.width = rect.width * 2; // 2x for retina
  canvas.height = rect.height * 2;

  const ctx = canvas.getContext('2d');
  if (!ctx) {
    console.error('Failed to get canvas 2d context');
    return;
  }

  // Try to find an SVG element inside the chart and render it
  const svgElement = chartRef.querySelector('svg');
  if (svgElement) {
    const svgData = new XMLSerializer().serializeToString(svgElement);
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);

    const img = new Image();
    img.onload = () => {
      ctx.scale(2, 2);
      ctx.drawImage(img, 0, 0, rect.width, rect.height);
      URL.revokeObjectURL(url);

      canvas.toBlob((blob) => {
        if (blob) {
          triggerDownload(blob, filename.endsWith('.png') ? filename : `${filename}.png`);
        }
      }, 'image/png');
    };
    img.onerror = () => {
      URL.revokeObjectURL(url);
      console.error('Failed to render SVG to canvas');
    };
    img.src = url;
  } else {
    // Fallback: export a blank canvas with background
    ctx.fillStyle = '#ffffff';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    ctx.fillStyle = '#333333';
    ctx.font = '24px sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('Chart Export', canvas.width / 2, canvas.height / 2);

    canvas.toBlob((blob) => {
      if (blob) {
        triggerDownload(blob, filename.endsWith('.png') ? filename : `${filename}.png`);
      }
    }, 'image/png');
  }
}

/**
 * Exports a chart element as an SVG file.
 * Extracts the SVG element from the chart container and triggers download.
 */
export function exportChartAsSVG(chartRef: HTMLElement, filename: string): void {
  const svgElement = chartRef.querySelector('svg');
  if (!svgElement) {
    console.error('No SVG element found in chart container');
    return;
  }

  // Clone the SVG to avoid modifying the original
  const clonedSvg = svgElement.cloneNode(true) as SVGElement;

  // Ensure the SVG has proper namespace and dimensions
  if (!clonedSvg.getAttribute('xmlns')) {
    clonedSvg.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
  }
  if (!clonedSvg.getAttribute('width')) {
    const rect = svgElement.getBoundingClientRect();
    clonedSvg.setAttribute('width', String(rect.width));
    clonedSvg.setAttribute('height', String(rect.height));
  }

  const svgData = new XMLSerializer().serializeToString(clonedSvg);
  const blob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });

  triggerDownload(blob, filename.endsWith('.svg') ? filename : `${filename}.svg`);
}

/**
 * Triggers a file download via a temporary anchor element.
 */
function triggerDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.style.display = 'none';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
