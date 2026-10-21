import type { CustomLayerInterface, CustomRenderMethodInput, Map as MapLibreMap } from 'maplibre-gl';
import { mercatorX, mercatorY } from '../geo.ts';

/** Straight (not premultiplied) RGBA, 0–1. */
export type Rgba = readonly [number, number, number, number];

export interface PointColors {
  selected: Rgba;
  context: Rgba;
}

const VERTEX = `#version 300 es
in vec2 a_pos;
in float a_state;
uniform mat4 u_matrix;
uniform float u_pass;
uniform float u_size;
void main() {
  if (abs(a_state - u_pass) > 0.5) {
    gl_Position = vec4(2.0, 2.0, 2.0, 1.0);
    gl_PointSize = 0.0;
    return;
  }
  gl_Position = u_matrix * vec4(a_pos, 0.0, 1.0);
  gl_PointSize = u_size;
}`;

const FRAGMENT = `#version 300 es
precision mediump float;
uniform vec4 u_color;
out vec4 fragColor;
void main() {
  vec2 c = gl_PointCoord * 2.0 - 1.0;
  float r = dot(c, c);
  if (r > 1.0) discard;
  float a = u_color.a * (1.0 - smoothstep(0.55, 1.0, r));
  fragColor = vec4(u_color.rgb * a, a);
}`;

/**
 * Draws every alert as a point, in one draw call per colour. Filtering only
 * rewrites a byte per point (2 = selected, 1 = context), so it's instant.
 *
 * Positions are stored relative to the middle of the data so they keep their
 * precision as 32-bit floats when zoomed in to street level.
 */
export class AlertPointsLayer implements CustomLayerInterface {
  readonly id: string;
  readonly type = 'custom';
  readonly renderingMode = '2d';

  private readonly count: number;
  private readonly originX: number;
  private readonly originY: number;
  /** Mercator positions relative to the origin, interleaved x, y. */
  private readonly positions: Float32Array;
  private state: Uint8Array;
  /** Brightens selected points when few are selected, so a narrow selection still shows up. */
  private boost = 1;
  private colors: PointColors;

  private map?: MapLibreMap;
  private gl?: WebGL2RenderingContext;
  private program?: WebGLProgram;
  private vao?: WebGLVertexArrayObject;
  private buffers: WebGLBuffer[] = [];
  private stateBuffer?: WebGLBuffer;
  private uniforms?: Record<'matrix' | 'pass' | 'size' | 'color', WebGLUniformLocation | null>;

  constructor(id: string, lon: Float64Array, lat: Float64Array, colors: PointColors) {
    this.id = id;
    this.count = lon.length;
    this.colors = colors;
    this.state = new Uint8Array(this.count).fill(2);

    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
    const mx = new Float64Array(this.count);
    const my = new Float64Array(this.count);
    for (let i = 0; i < this.count; i++) {
      mx[i] = mercatorX(lon[i]);
      my[i] = mercatorY(lat[i]);
      minX = Math.min(minX, mx[i]);
      maxX = Math.max(maxX, mx[i]);
      minY = Math.min(minY, my[i]);
      maxY = Math.max(maxY, my[i]);
    }
    this.originX = (minX + maxX) / 2;
    this.originY = (minY + maxY) / 2;
    this.positions = new Float32Array(this.count * 2);
    for (let i = 0; i < this.count; i++) {
      this.positions[2 * i] = mx[i] - this.originX;
      this.positions[2 * i + 1] = my[i] - this.originY;
    }
  }

  setState(state: Uint8Array): void {
    this.state = state;
    let selected = 0;
    for (let i = 0; i < state.length; i++) if (state[i] === 2) selected++;
    this.boost = clamp(Math.sqrt(this.count / Math.max(1, selected)), 1, 3);
    const gl = this.gl;
    if (gl && this.stateBuffer) {
      gl.bindBuffer(gl.ARRAY_BUFFER, this.stateBuffer);
      gl.bufferSubData(gl.ARRAY_BUFFER, 0, state);
    }
    this.map?.triggerRepaint();
  }

  setColors(colors: PointColors): void {
    this.colors = colors;
    this.map?.triggerRepaint();
  }

  /**
   * The alert nearest to a mercator position, within `radius` (mercator
   * units), among those with the given state. -1 if none.
   */
  nearest(x: number, y: number, radius: number, state = 2): number {
    const px = x - this.originX;
    const py = y - this.originY;
    let best = -1;
    let bestD = radius * radius;
    for (let i = 0; i < this.count; i++) {
      if (this.state[i] !== state) continue;
      const dx = this.positions[2 * i] - px;
      const dy = this.positions[2 * i + 1] - py;
      const d = dx * dx + dy * dy;
      if (d <= bestD) {
        bestD = d;
        best = i;
      }
    }
    return best;
  }

  onAdd(map: MapLibreMap, gl: WebGL2RenderingContext): void {
    this.map = map;
    this.gl = gl;
    const program = link(gl, compile(gl, gl.VERTEX_SHADER, VERTEX), compile(gl, gl.FRAGMENT_SHADER, FRAGMENT));
    this.program = program;
    this.uniforms = {
      matrix: gl.getUniformLocation(program, 'u_matrix'),
      pass: gl.getUniformLocation(program, 'u_pass'),
      size: gl.getUniformLocation(program, 'u_size'),
      color: gl.getUniformLocation(program, 'u_color'),
    };

    const vao = gl.createVertexArray();
    gl.bindVertexArray(vao);

    const posBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, posBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, this.positions, gl.STATIC_DRAW);
    const aPos = gl.getAttribLocation(program, 'a_pos');
    gl.enableVertexAttribArray(aPos);
    gl.vertexAttribPointer(aPos, 2, gl.FLOAT, false, 0, 0);

    const stateBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, stateBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, this.state, gl.DYNAMIC_DRAW);
    const aState = gl.getAttribLocation(program, 'a_state');
    gl.enableVertexAttribArray(aState);
    gl.vertexAttribPointer(aState, 1, gl.UNSIGNED_BYTE, false, 0, 0);

    gl.bindVertexArray(null);
    this.vao = vao;
    this.buffers = [posBuffer, stateBuffer];
    this.stateBuffer = stateBuffer;
  }

  onRemove(_map: MapLibreMap, gl: WebGL2RenderingContext): void {
    for (const b of this.buffers) gl.deleteBuffer(b);
    if (this.vao) gl.deleteVertexArray(this.vao);
    if (this.program) gl.deleteProgram(this.program);
    this.buffers = [];
    this.vao = this.program = this.stateBuffer = this.uniforms = this.gl = undefined;
  }

  render(gl: WebGL2RenderingContext, options: CustomRenderMethodInput): void {
    if (!this.program || !this.vao || !this.uniforms || !this.map) return;

    // Fold the origin offset into the mercator-to-clip matrix (column-major).
    const m = options.defaultProjectionData.mainMatrix;
    const matrix = new Float32Array(16);
    for (let i = 0; i < 12; i++) matrix[i] = m[i];
    for (let r = 0; r < 4; r++) matrix[12 + r] = m[r] * this.originX + m[4 + r] * this.originY + m[12 + r];

    // Points grow and firm up as you zoom in, so dense roads read as lines
    // when zoomed out and single alerts stay visible when zoomed in.
    const zoom = this.map.getZoom();
    const size = clamp(1.6 + (zoom - 7) * 0.9, 1.6, 11) * devicePixelRatio;
    const strength = clamp(0.22 + (zoom - 7) * 0.09, 0.22, 0.9);

    gl.useProgram(this.program);
    gl.bindVertexArray(this.vao);
    gl.uniformMatrix4fv(this.uniforms.matrix, false, matrix);
    gl.uniform1f(this.uniforms.size, size);
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA);

    const passes: [number, Rgba, number][] = [
      [1, this.colors.context, strength * 0.7],
      [2, this.colors.selected, Math.min(0.95, strength * this.boost)],
    ];
    for (const [pass, [r, g, b, a], k] of passes) {
      gl.uniform1f(this.uniforms.pass, pass);
      gl.uniform4f(this.uniforms.color, r, g, b, a * k);
      gl.drawArrays(gl.POINTS, 0, this.count);
    }
    gl.bindVertexArray(null);
  }
}

function clamp(v: number, lo: number, hi: number): number {
  return Math.min(hi, Math.max(lo, v));
}

function compile(gl: WebGL2RenderingContext, type: number, source: string): WebGLShader {
  const shader = gl.createShader(type)!;
  gl.shaderSource(shader, source);
  gl.compileShader(shader);
  if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
    throw new Error(`Shader failed to compile: ${gl.getShaderInfoLog(shader)}`);
  }
  return shader;
}

function link(gl: WebGL2RenderingContext, vertex: WebGLShader, fragment: WebGLShader): WebGLProgram {
  const program = gl.createProgram();
  gl.attachShader(program, vertex);
  gl.attachShader(program, fragment);
  gl.linkProgram(program);
  gl.deleteShader(vertex);
  gl.deleteShader(fragment);
  if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
    throw new Error(`Shader program failed to link: ${gl.getProgramInfoLog(program)}`);
  }
  return program;
}
