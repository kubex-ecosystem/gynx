import * as React from 'react';

import { AlertTriangle, Loader } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import * as Loader2 from '../common/LoaderAlt';

interface MermaidDiagramProps {
  chart: string;
}

// Removido top-level await e componente async. Carregamos mermaid dentro do hook.
const MermaidDiagram: React.FC<MermaidDiagramProps> = ({ chart }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const mmdRef = useRef<any | null>(null);
  const [svg, setSvg] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Generate a unique ID for each diagram container
  // ...existing code...
  let diagramIdCounter = 0;
  const [uniqueId] = useState(() => `mermaid-diagram-${diagramIdCounter++}`);

  // Load and initialize mermaid once on mount
  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    Loader2.loadMermaidFromCDN()
      .then((mod) => {
        if (cancelled) return;
        mmdRef.current = mod;
        try {
          mmdRef.current.initialize({
            startOnLoad: false,
            theme: 'dark',
            securityLevel: 'loose',
            fontFamily: 'Inter, sans-serif',
            themeVariables: {
              background: '#1f2937', // gray-800
              primaryColor: '#374151', // gray-700
              primaryTextColor: '#f3f4f6', // gray-100
              lineColor: '#a78bfa', // purple-400
              textColor: '#d1d5db', // gray-300
            },
          });
        } catch (initErr) {
          console.error("Mermaid initialize error:", initErr);
          setError('Failed to initialize diagram renderer.');
        }
      })
      .catch((err) => {
        console.error('Failed to load Mermaid:', err);
        setError('Failed to load diagram renderer.');
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  // Render the chart when mermaid is ready or when chart changes
  useEffect(() => {
    if (!mmdRef.current) return;
    if (!chart) {
      setSvg(null);
      return;
    }

    setIsLoading(true);
    setError(null);
    let cancelled = false;

    mmdRef.current
      .render(uniqueId, chart)
      .then((result: any) => {
        if (cancelled) return;
        // mermaid.render may return an object with svg or a string depending on version
        const svgOutput =
          result && typeof result === 'object' && 'svg' in result
            ? result.svg
            : String(result || '');
        setSvg(svgOutput);
      })
      .catch((err: any) => {
        console.error('Mermaid render error:', err);
        setError('Failed to render the diagram. The generated syntax might be invalid.');
        setSvg(null);
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
    // Não incluir 'svg' nos deps para evitar loop; depende de chart e da instância carregada
  }, [chart, uniqueId]);

  return (
    <div className="p-4 bg-gray-900/50 border border-gray-700 rounded-lg min-h-[200px] flex items-center justify-center">
      {isLoading && <Loader className="w-8 h-8 text-purple-400 animate-spin" />}
      {error && (
        <div className="text-center text-red-400">
          <AlertTriangle className="w-8 h-8 mx-auto mb-2" />
          <p>{error}</p>
        </div>
      )}
      {svg && !isLoading && (
        <div
          ref={containerRef}
          dangerouslySetInnerHTML={{ __html: svg }}
          className="w-full h-full flex items-center justify-center"
        />
      )}
    </div>
  );
};

export default MermaidDiagram;
